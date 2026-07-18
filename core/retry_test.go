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

// flakyRangeServer supports Range but returns 500 the first time each distinct
// range is requested, then succeeds — so only retry logic can complete it.
func flakyRangeServer(body []byte) *httptest.Server {
	var mu sync.Mutex
	seen := map[string]int{}
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodHead {
			w.Header().Set("Accept-Ranges", "bytes")
			w.Header().Set("Content-Length", itoa(len(body)))
			w.WriteHeader(http.StatusOK)
			return
		}
		rng := r.Header.Get("Range")
		mu.Lock()
		n := seen[rng]
		seen[rng]++
		mu.Unlock()
		if rng != "" && n == 0 {
			http.Error(w, "flaky", http.StatusInternalServerError)
			return
		}
		start, end := 0, len(body)-1
		if rng != "" {
			fmt.Sscanf(rng, "bytes=%d-%d", &start, &end)
		}
		w.Header().Set("Accept-Ranges", "bytes")
		if rng != "" {
			w.Header().Set("Content-Range", fmt.Sprintf("bytes %d-%d/%d", start, end, len(body)))
			w.Header().Set("Content-Length", itoa(end-start+1))
			w.WriteHeader(http.StatusPartialContent)
		} else {
			w.Header().Set("Content-Length", itoa(len(body)))
			w.WriteHeader(http.StatusOK)
		}
		w.Write(body[start : end+1])
	}))
}

func TestSectionRetrySucceeds(t *testing.T) {
	body := makeBody(20000)
	srv := flakyRangeServer(body)
	defer srv.Close()

	m := newTestManager(t)
	id, err := m.Add(srv.URL + "/f.bin")
	if err != nil {
		t.Fatal(err)
	}
	waitFor(t, func() bool {
		s := stateOf(m, id)
		if s == StateFailed {
			t.Fatal("download failed instead of retrying")
		}
		return s == StateDone
	}, 15*time.Second)

	got, _ := os.ReadFile(m.findJob(id).TargetPath)
	if sha256.Sum256(got) != sha256.Sum256(body) {
		t.Fatal("content mismatch after retries")
	}
}
