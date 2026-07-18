package core

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"

	"simple-gui/helper"
)

// Manager owns all download state and is the single source of truth shared by
// both front-ends (the TUI and the giu GUI). It is safe for concurrent use.
type Manager struct {
	mu              sync.Mutex
	jobs            []*Job
	downloadDir     string
	stateDir        string
	maxActive       int   // cap on simultaneously-downloading jobs
	bytesPerSection int64 // target bytes per parallel section
	notify          bool  // fire a macOS notification on completion

	inboxMu  sync.Mutex
	done     chan struct{}
	closeOne sync.Once
}

const (
	defaultMaxActive       = 3
	defaultBytesPerSection = 2 * 1024 * 1024
)

// NewManager loads config, resolves the download/state directories, creates
// them, and returns a ready manager. (Folds the old core.Start dir-setup.)
func NewManager() (*Manager, error) {
	cfg := LoadConfig()
	st, err := GetTempPath()
	if err != nil {
		return nil, err
	}
	m, err := newManager(cfg.DownloadDir, st)
	if err != nil {
		return nil, err
	}
	m.maxActive = cfg.MaxActive
	m.bytesPerSection = int64(cfg.SectionSizeMB) * 1024 * 1024
	m.notify = cfg.Notifications
	return m, nil
}

// newManager is the injectable constructor used by tests.
func newManager(downloadDir, stateDir string) (*Manager, error) {
	if err := os.MkdirAll(downloadDir, os.ModePerm); err != nil {
		return nil, err
	}
	if err := os.MkdirAll(stateDir, os.ModePerm); err != nil {
		return nil, err
	}
	m := &Manager{downloadDir: downloadDir, stateDir: stateDir, maxActive: defaultMaxActive, bytesPerSection: defaultBytesPerSection, done: make(chan struct{})}
	go m.sampleLoop()
	go m.inboxLoop()
	return m, nil
}

// schedule promotes queued jobs to downloading while a slot is free. A queued
// job is claimed (state set to downloading) under the lock before its goroutine
// launches, so concurrent schedule calls never double-start the same job.
func (m *Manager) schedule() {
	m.mu.Lock()
	defer m.mu.Unlock()
	active := 0
	for _, j := range m.jobs {
		switch j.getState() {
		case StateDownloading, StateMerging:
			active++
		}
	}
	for _, j := range m.jobs {
		if active >= m.maxActive {
			break
		}
		if j.getState() == StateQueued {
			j.setState(StateDownloading)
			active++
			go m.start(j)
		}
	}
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

	totalSection := 1
	if details.AcceptRanges && details.Size > 0 {
		totalSection = sectionsForSize(details.Size, m.bytesPerSection)
	}

	m.mu.Lock()
	name := m.uniqueFilenameLocked(details.Name)
	j := &Job{
		ID:           uuid.NewString(),
		URL:          rawURL,
		Filename:     name,
		TargetPath:   filepath.Join(m.downloadDir, name),
		Size:         details.Size,
		AcceptRanges: details.AcceptRanges,
		TotalSection: totalSection,
		state:        StateQueued,
	}
	if j.isSegmented() {
		j.Sections = computeSections(details.Size, j.TotalSection)
	}
	m.jobs = append(m.jobs, j)
	m.mu.Unlock()

	if err := writeMeta(m.stateDir, m.metaFor(j)); err != nil {
		return "", err
	}

	m.schedule() // starts now if under the cap, otherwise leaves it queued
	return j.ID, nil
}

// uniqueFilenameLocked returns name if free, otherwise "base (n).ext" with the
// smallest n that avoids both an existing file on disk and any active job's
// filename. Caller must hold m.mu.
func (m *Manager) uniqueFilenameLocked(name string) string {
	ext := filepath.Ext(name)
	base := name[:len(name)-len(ext)]
	candidate := name
	for n := 1; m.nameTakenLocked(candidate); n++ {
		candidate = fmt.Sprintf("%s (%d)%s", base, n, ext)
	}
	return candidate
}

func (m *Manager) nameTakenLocked(name string) bool {
	if fileExists(filepath.Join(m.downloadDir, name)) {
		return true
	}
	for _, j := range m.jobs {
		if j.Filename == name {
			return true
		}
	}
	return false
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
	if j.startedAt.IsZero() {
		j.startedAt = time.Now() // first start only; preserved across resume
	}
	j.mu.Unlock()

	err := j.run(ctx, m.stateDir)

	completed := false
	j.mu.Lock()
	switch {
	case errors.Is(err, errPaused):
		j.state = StatePaused
	case err != nil:
		j.state = StateFailed
		j.err = err
	default:
		j.state = StateDone
		completed = true
	}
	name := j.Filename
	j.mu.Unlock()

	if completed && m.notify {
		go notifyDone(name)
	}
	m.schedule() // a slot freed — promote the next queued job
}

// notifyDone posts a macOS notification that a download finished.
func notifyDone(name string) {
	safe := strings.NewReplacer(`\`, `\\`, `"`, `\"`).Replace(name)
	script := fmt.Sprintf(`display notification "%s" with title "godownloader" subtitle "Download complete"`, safe)
	exec.Command("osascript", "-e", script).Run()
}

// Snapshot returns an immutable, point-in-time copy of current job state for
// rendering. The lock is held throughout so the view is consistent with the
// scheduler (e.g. the number of downloading jobs never appears to exceed the
// cap due to a promotion happening mid-iteration).
func (m *Manager) Snapshot() []JobStatus {
	m.mu.Lock()
	defer m.mu.Unlock()
	out := make([]JobStatus, len(m.jobs))
	for i, j := range m.jobs {
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
	if canResume {
		j.state = StateQueued // re-enter the queue; schedule respects the cap
	}
	j.mu.Unlock()
	if !canResume {
		return
	}
	m.schedule()
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
		// Done only if the final file is present, complete (full size when
		// known), and no partial temp files remain. Otherwise it is resumable.
		complete := fileExists(j.TargetPath) && !j.hasTempFiles(m.stateDir) &&
			(j.Size == 0 || fileSize(j.TargetPath) >= j.Size)
		if complete {
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

	m.schedule() // removing an active job frees a slot
}

// ClearCompleted removes finished downloads from the list and deletes their
// metadata (the downloaded files are kept).
func (m *Manager) ClearCompleted() {
	m.mu.Lock()
	kept := m.jobs[:0]
	var removed []*Job
	for _, j := range m.jobs {
		if j.getState() == StateDone {
			removed = append(removed, j)
		} else {
			kept = append(kept, j)
		}
	}
	m.jobs = kept
	m.mu.Unlock()

	for _, j := range removed {
		os.Remove(metaPath(m.stateDir, j.ID))
	}
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

// OpenFolder reveals the downloads directory in the OS file browser (macOS).
func (m *Manager) OpenFolder() error {
	return exec.Command("open", m.downloadDir).Run()
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
