package core

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type JobState string

const (
	StateQueued      JobState = "queued"
	StateDownloading JobState = "downloading"
	StatePaused      JobState = "paused"
	StateMerging     JobState = "merging"
	StateDone        JobState = "done"
	StateFailed      JobState = "failed"
)

// errPaused is returned by run when the context was cancelled by a Pause, so
// the manager can distinguish a deliberate pause from a real failure.
var errPaused = errors.New("paused")

// Job is one download. It owns its section goroutines; bytes-downloaded is not
// stored here but derived from temp file sizes on disk (see snapshot).
type Job struct {
	ID, URL, Filename, TargetPath string
	Size                          int64
	AcceptRanges                  bool
	Sections                      [][2]int64
	TotalSection                  int

	mu             sync.Mutex
	state          JobState
	err            error
	cancel         context.CancelFunc
	speed          float64 // EWMA bytes/sec
	lastBytes      int64
	lastSampleTime time.Time
}

// JobStatus is an immutable snapshot of a Job for rendering by the front-ends.
type JobStatus struct {
	ID, Name         string
	Size, Downloaded int64
	Percent, Speed   float64
	ETASeconds       int
	State            JobState
	Resumable        bool
	Err              string
}

func (j *Job) setState(s JobState) {
	j.mu.Lock()
	j.state = s
	j.mu.Unlock()
}

func (j *Job) getState() JobState {
	j.mu.Lock()
	defer j.mu.Unlock()
	return j.state
}

// isSegmented reports whether the job can be downloaded in parallel sections
// (and therefore resumed via Range requests).
func (j *Job) isSegmented() bool {
	return j.AcceptRanges && j.Size > 0 && j.TotalSection > 1
}

// statusWithDownloaded builds the render DTO from a supplied downloaded count.
// Kept pure for unit testing; the manager uses snapshot for locked reads.
func (j *Job) statusWithDownloaded(downloaded int64) JobStatus {
	var pct float64
	if j.Size > 0 {
		pct = float64(downloaded) / float64(j.Size) * 100
	}
	errStr := ""
	if j.err != nil {
		errStr = j.err.Error()
	}
	return JobStatus{
		ID:         j.ID,
		Name:       j.Filename,
		Size:       j.Size,
		Downloaded: downloaded,
		Percent:    pct,
		Speed:      j.speed,
		ETASeconds: etaSeconds(j.Size-downloaded, j.speed),
		State:      j.state,
		Resumable:  j.AcceptRanges && j.Size > 0,
		Err:        errStr,
	}
}

// measured returns the bytes downloaded so far, derived from disk (temp files
// for segmented jobs, the target file otherwise). Pure w.r.t. job locks.
func (j *Job) measured(tempDir string, state JobState) int64 {
	switch {
	case state == StateDone:
		return j.Size
	case j.isSegmented():
		var total int64
		for i := range j.Sections {
			total += fileSize(j.sectionFile(tempDir, i))
		}
		return total
	default:
		return fileSize(j.TargetPath)
	}
}

// snapshot reads job state under lock, then measures downloaded bytes from disk.
func (j *Job) snapshot(tempDir string) JobStatus {
	j.mu.Lock()
	state, speed := j.state, j.speed
	var errStr string
	if j.err != nil {
		errStr = j.err.Error()
	}
	j.mu.Unlock()

	downloaded := j.measured(tempDir, state)

	var pct float64
	if j.Size > 0 {
		pct = float64(downloaded) / float64(j.Size) * 100
	}
	return JobStatus{
		ID:         j.ID,
		Name:       j.Filename,
		Size:       j.Size,
		Downloaded: downloaded,
		Percent:    pct,
		Speed:      speed,
		ETASeconds: etaSeconds(j.Size-downloaded, speed),
		State:      state,
		Resumable:  j.AcceptRanges && j.Size > 0,
		Err:        errStr,
	}
}

func (j *Job) sectionFile(tempDir string, i int) string {
	return filepath.Join(tempDir, fmt.Sprintf("section-%d-%s.tmp", i+1, j.Filename))
}

// run downloads the job to completion (or returns errPaused on cancel). It is
// safe to call again to resume: each section continues from the bytes already
// present in its temp file.
func (j *Job) run(ctx context.Context, tempDir string) error {
	if !j.isSegmented() {
		return j.singleStream(ctx)
	}

	wg := sync.WaitGroup{}
	errs := make([]error, len(j.Sections))
	for i, sec := range j.Sections {
		have := fileSize(j.sectionFile(tempDir, i))
		start, end, ok := remainingRange(sec, have)
		if !ok {
			continue // section already complete
		}
		truncate := have == 0
		wg.Add(1)
		go func(i int, start, end int64, truncate bool) {
			defer wg.Done()
			errs[i] = j.downloadSection(ctx, tempDir, i, start, end, truncate)
		}(i, start, end, truncate)
	}
	wg.Wait()

	if ctx.Err() != nil {
		return errPaused // paused; keep temp files for resume
	}
	for _, e := range errs {
		if e != nil {
			return e
		}
	}

	j.setState(StateMerging)
	if err := j.mergeFiles(tempDir); err != nil {
		return err
	}
	j.cleanupTempFiles(tempDir)
	return nil
}

func (j *Job) downloadSection(ctx context.Context, tempDir string, i int, start, end int64, truncate bool) error {
	req, err := newRequest(j.URL, "GET")
	if err != nil {
		return err
	}
	req.Header.Set("Range", fmt.Sprintf("bytes=%d-%d", start, end))
	resp, err := httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// If the server ignored the Range header it returns 200 with the FULL body;
	// writing that into a section would corrupt the merge.
	if resp.StatusCode != http.StatusPartialContent {
		return fmt.Errorf("section %d: server did not honor range request (status %d)", i+1, resp.StatusCode)
	}

	flags := os.O_CREATE | os.O_WRONLY | os.O_APPEND
	if truncate {
		flags = os.O_CREATE | os.O_WRONLY | os.O_TRUNC
	}
	f, err := os.OpenFile(j.sectionFile(tempDir, i), flags, os.ModePerm)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = io.Copy(f, &ctxReader{ctx: ctx, r: resp.Body})
	if err != nil && ctx.Err() != nil {
		return nil // cancelled: bytes written so far are a valid resume point
	}
	return err
}

// singleStream downloads the whole body to the target in one request (used when
// the server lacks Range support or the size is unknown). Not resumable.
func (j *Job) singleStream(ctx context.Context) error {
	req, err := newRequest(j.URL, "GET")
	if err != nil {
		return err
	}
	resp, err := httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode > 299 {
		return fmt.Errorf("status code: %d", resp.StatusCode)
	}

	f, err := os.OpenFile(j.TargetPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, os.ModePerm)
	if err != nil {
		return err
	}
	defer f.Close()

	if _, err := io.Copy(f, &ctxReader{ctx: ctx, r: resp.Body}); err != nil {
		if ctx.Err() != nil {
			return errPaused
		}
		return err
	}
	return nil
}

func (j *Job) mergeFiles(tempDir string) error {
	// O_TRUNC so a stale file at the path is overwritten, not appended to.
	f, err := os.OpenFile(j.TargetPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, os.ModePerm)
	if err != nil {
		return err
	}
	defer f.Close()

	for i := range j.Sections {
		sf, err := os.Open(j.sectionFile(tempDir, i))
		if err != nil {
			return err
		}
		_, err = io.Copy(f, sf)
		sf.Close()
		if err != nil {
			return err
		}
	}
	return nil
}

func (j *Job) cleanupTempFiles(tempDir string) {
	for i := range j.Sections {
		os.Remove(j.sectionFile(tempDir, i))
	}
}

// ctxReader aborts a Read as soon as its context is cancelled, so io.Copy stops
// promptly when a download is paused.
type ctxReader struct {
	ctx context.Context
	r   io.Reader
}

func (c *ctxReader) Read(p []byte) (int, error) {
	if err := c.ctx.Err(); err != nil {
		return 0, err
	}
	return c.r.Read(p)
}

func fileSize(path string) int64 {
	fi, err := os.Stat(path)
	if err != nil {
		return 0
	}
	return fi.Size()
}
