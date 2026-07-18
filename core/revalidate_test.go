package core

import (
	"crypto/sha256"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"sync"
	"testing"
	"time"
)

// versionedServer serves a body + ETag that can be swapped mid-test, with slow
// Range support so a download can be paused between versions.
type versionedServer struct {
	*httptest.Server
	mu    sync.Mutex
	body  []byte
	etag  string
	chunk int
	delay time.Duration
}

func newVersionedServer(body []byte, etag string, chunk int, delay time.Duration) *versionedServer {
	vs := &versionedServer{body: body, etag: etag, chunk: chunk, delay: delay}
	vs.Server = httptest.NewServer(http.HandlerFunc(vs.handle))
	return vs
}

func (vs *versionedServer) swap(body []byte, etag string) {
	vs.mu.Lock()
	vs.body, vs.etag = body, etag
	vs.mu.Unlock()
}

func (vs *versionedServer) handle(w http.ResponseWriter, r *http.Request) {
	vs.mu.Lock()
	body, etag, chunk, delay := vs.body, vs.etag, vs.chunk, vs.delay
	vs.mu.Unlock()

	w.Header().Set("Accept-Ranges", "bytes")
	w.Header().Set("ETag", etag)

	start, end := 0, len(body)-1
	hasRange := r.Header.Get("Range") != ""
	if hasRange {
		fmt.Sscanf(r.Header.Get("Range"), "bytes=%d-%d", &start, &end)
	}
	if hasRange {
		w.Header().Set("Content-Range", fmt.Sprintf("bytes %d-%d/%d", start, end, len(body)))
		w.Header().Set("Content-Length", itoa(end-start+1))
		w.WriteHeader(http.StatusPartialContent)
	} else {
		w.Header().Set("Content-Length", itoa(len(body)))
		w.WriteHeader(http.StatusOK)
	}
	if r.Method == http.MethodHead {
		return
	}
	flusher, _ := w.(http.Flusher)
	for i := start; i <= end; i += chunk {
		j := i + chunk - 1
		if j > end {
			j = end
		}
		w.Write(body[i : j+1])
		if flusher != nil {
			flusher.Flush()
		}
		time.Sleep(delay)
	}
}

// TestResumeRestartsWhenFileChanged: download partway, pause, swap the server's
// body+ETag, resume — the resulting file must be the NEW body (not a corrupt
// mix of old + new bytes).
func TestResumeRestartsWhenFileChanged(t *testing.T) {
	bodyA := makeBody(100000)
	bodyB := make([]byte, 100000)
	for i := range bodyB {
		bodyB[i] = byte((i + 7) % 251) // different content, same length
	}
	srv := newVersionedServer(bodyA, `"v1"`, 512, 12*time.Millisecond)
	defer srv.Close()

	m := newTestManager(t)
	id, err := m.Add(srv.URL + "/f.bin")
	if err != nil {
		t.Fatal(err)
	}
	waitFor(t, func() bool {
		s, _ := statusOf(m, id)
		return s.Downloaded > 0 && s.Downloaded < s.Size
	}, 3*time.Second)
	m.Pause(id)
	waitFor(t, func() bool { return stateOf(m, id) == StatePaused }, 3*time.Second)

	// The file changes on the server while paused.
	srv.swap(bodyB, `"v2"`)

	m.Resume(id)
	waitFor(t, func() bool { return stateOf(m, id) == StateDone }, 10*time.Second)

	got, _ := os.ReadFile(m.findJob(id).TargetPath)
	if sha256.Sum256(got) != sha256.Sum256(bodyB) {
		t.Fatal("resumed file does not match the new server content (stale temp files were not discarded)")
	}
}
