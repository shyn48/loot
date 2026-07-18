package core

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/google/uuid"

	"simple-gui/helper"
)

// Manager owns all download state and is the single source of truth shared by
// both front-ends (the TUI and the giu GUI). It is safe for concurrent use.
type Manager struct {
	mu          sync.Mutex
	jobs        []*Job
	downloadDir string
	stateDir    string

	done     chan struct{}
	closeOne sync.Once
}

// NewManager resolves the standard download/temp directories, creates them, and
// returns a ready manager. (Folds the old core.Start dir-setup.)
func NewManager() (*Manager, error) {
	dl, err := GetDownloadPath("")
	if err != nil {
		return nil, err
	}
	st, err := GetTempPath()
	if err != nil {
		return nil, err
	}
	return newManager(dl, st)
}

// newManager is the injectable constructor used by tests.
func newManager(downloadDir, stateDir string) (*Manager, error) {
	if err := os.MkdirAll(downloadDir, os.ModePerm); err != nil {
		return nil, err
	}
	if err := os.MkdirAll(stateDir, os.ModePerm); err != nil {
		return nil, err
	}
	m := &Manager{downloadDir: downloadDir, stateDir: stateDir, done: make(chan struct{})}
	go m.sampleLoop()
	return m, nil
}

// sampleLoop periodically updates each running job's transfer speed until Close.
func (m *Manager) sampleLoop() {
	t := time.NewTicker(500 * time.Millisecond)
	defer t.Stop()
	for {
		select {
		case <-m.done:
			return
		case <-t.C:
			m.sample()
		}
	}
}

func (m *Manager) sample() {
	m.mu.Lock()
	jobs := make([]*Job, len(m.jobs))
	copy(jobs, m.jobs)
	m.mu.Unlock()

	now := time.Now()
	for _, j := range jobs {
		if j.getState() != StateDownloading {
			continue
		}
		downloaded := j.measured(m.stateDir, StateDownloading)
		j.mu.Lock()
		if dt := now.Sub(j.lastSampleTime).Seconds(); dt > 0 {
			s := float64(downloaded-j.lastBytes) / dt
			if s < 0 {
				s = 0
			}
			j.speed = ewma(j.speed, s, 0.3)
		}
		j.lastBytes = downloaded
		j.lastSampleTime = now
		j.mu.Unlock()
	}
}

// Close stops the manager's background goroutines.
func (m *Manager) Close() {
	m.closeOne.Do(func() { close(m.done) })
}

// Add probes the URL, creates a job, persists its metadata, and starts
// downloading in the background. It returns immediately with the job id.
func (m *Manager) Add(rawURL string) (string, error) {
	if !helper.IsValidUrl(rawURL) {
		return "", fmt.Errorf("invalid url: %s", rawURL)
	}
	details, err := GetFileDetails(rawURL)
	if err != nil {
		return "", err
	}

	j := &Job{
		ID:           uuid.NewString(),
		URL:          rawURL,
		Filename:     details.Name,
		TargetPath:   filepath.Join(m.downloadDir, details.Name),
		Size:         details.Size,
		AcceptRanges: details.AcceptRanges,
		TotalSection: 20,
		state:        StateQueued,
	}
	if j.isSegmented() {
		j.Sections = computeSections(details.Size, j.TotalSection)
	}
	if err := writeMeta(m.stateDir, m.metaFor(j)); err != nil {
		return "", err
	}

	m.mu.Lock()
	m.jobs = append(m.jobs, j)
	m.mu.Unlock()

	go m.start(j)
	return j.ID, nil
}

// start runs a job to completion and records its terminal state.
func (m *Manager) start(j *Job) {
	ctx, cancel := context.WithCancel(context.Background())
	seed := j.measured(m.stateDir, StateDownloading) // may be >0 when resuming
	j.mu.Lock()
	j.cancel = cancel
	j.state = StateDownloading
	j.lastBytes = seed
	j.lastSampleTime = time.Now()
	j.mu.Unlock()

	err := j.run(ctx, m.stateDir)

	j.mu.Lock()
	switch {
	case errors.Is(err, errPaused):
		j.state = StatePaused
	case err != nil:
		j.state = StateFailed
		j.err = err
	default:
		j.state = StateDone
	}
	j.mu.Unlock()
}

// Snapshot returns an immutable copy of current job state for rendering.
func (m *Manager) Snapshot() []JobStatus {
	m.mu.Lock()
	jobs := make([]*Job, len(m.jobs))
	copy(jobs, m.jobs)
	m.mu.Unlock()

	out := make([]JobStatus, len(jobs))
	for i, j := range jobs {
		out[i] = j.snapshot(m.stateDir)
	}
	return out
}

// Pause cancels a running download; its partial temp files remain so Resume can
// continue from where it stopped.
func (m *Manager) Pause(id string) {
	j := m.findJob(id)
	if j == nil {
		return
	}
	// Only cancel here; the state flips to Paused in start() once run() returns,
	// which guarantees every section goroutine has actually stopped writing.
	j.mu.Lock()
	if j.state == StateDownloading && j.cancel != nil {
		j.cancel()
	}
	j.mu.Unlock()
}

// Resume restarts a paused (or resumable-failed) download. Because run() derives
// each section's progress from its temp file size and requests only the missing
// range, resume needs no special logic — it just re-enters start(). A
// non-resumable job restarts from zero (single-stream truncates on re-run).
func (m *Manager) Resume(id string) {
	j := m.findJob(id)
	if j == nil {
		return
	}
	j.mu.Lock()
	resumable := j.AcceptRanges && j.Size > 0
	canResume := j.state == StatePaused || (j.state == StateFailed && resumable)
	j.mu.Unlock()
	if !canResume {
		return
	}
	go m.start(j)
}

// LoadPersisted rebuilds jobs from metadata left on disk by a previous run.
// Incomplete downloads come back Paused (the user resumes them); already-
// finished ones come back Done. Nothing auto-starts.
func (m *Manager) LoadPersisted() error {
	files, err := listMetaFiles(m.stateDir)
	if err != nil {
		return err
	}
	for _, f := range files {
		md, err := readMeta(f)
		if err != nil {
			continue
		}
		j := &Job{
			ID:           md.ID,
			URL:          md.URL,
			Filename:     md.Filename,
			TargetPath:   md.TargetPath,
			Size:         md.Size,
			AcceptRanges: md.AcceptRanges,
			Sections:     md.Sections,
			TotalSection: md.TotalSection,
		}
		if fileExists(j.TargetPath) && !j.hasTempFiles(m.stateDir) {
			j.state = StateDone
		} else {
			j.state = StatePaused
		}
		m.mu.Lock()
		m.jobs = append(m.jobs, j)
		m.mu.Unlock()
	}
	return nil
}

// Remove cancels a job (if running) and deletes its temp files and metadata. A
// completed download's final file is kept.
func (m *Manager) Remove(id string) {
	m.mu.Lock()
	var target *Job
	for i, j := range m.jobs {
		if j.ID == id {
			target = j
			m.jobs = append(m.jobs[:i], m.jobs[i+1:]...)
			break
		}
	}
	m.mu.Unlock()
	if target == nil {
		return
	}

	target.mu.Lock()
	if target.cancel != nil {
		target.cancel()
	}
	state := target.state
	target.mu.Unlock()

	target.cleanupTempFiles(m.stateDir)
	if state != StateDone && !target.isSegmented() {
		os.Remove(target.TargetPath) // partial single-stream file
	}
	os.Remove(metaPath(m.stateDir, id))
}

// PauseAll pauses every running download and waits briefly for them to stop, so
// their temp files settle before the process exits (used on TUI quit).
func (m *Manager) PauseAll() {
	m.mu.Lock()
	jobs := make([]*Job, len(m.jobs))
	copy(jobs, m.jobs)
	m.mu.Unlock()
	for _, j := range jobs {
		m.Pause(j.ID)
	}
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if !m.anyDownloading() {
			return
		}
		time.Sleep(20 * time.Millisecond)
	}
}

func (m *Manager) anyDownloading() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, j := range m.jobs {
		if j.getState() == StateDownloading {
			return true
		}
	}
	return false
}

func (m *Manager) metaFor(j *Job) meta {
	return meta{
		ID:           j.ID,
		URL:          j.URL,
		Filename:     j.Filename,
		TargetPath:   j.TargetPath,
		Size:         j.Size,
		TotalSection: j.TotalSection,
		Sections:     j.Sections,
		AcceptRanges: j.AcceptRanges,
		CreatedAt:    time.Now(),
	}
}

func (m *Manager) findJob(id string) *Job {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, j := range m.jobs {
		if j.ID == id {
			return j
		}
	}
	return nil
}
