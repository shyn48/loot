package core

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// slowRangeServer supports Range requests but streams the response in small
// chunks with a delay, so a download can be observed and paused mid-flight.
func slowRangeServer(body []byte, chunk int, delay time.Duration) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start, end := 0, len(body)-1
		hasRange := r.Header.Get("Range") != ""
		if hasRange {
			fmt.Sscanf(r.Header.Get("Range"), "bytes=%d-%d", &start, &end)
		}
		w.Header().Set("Accept-Ranges", "bytes")
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
	}))
}

func TestPauseStopsDownload(t *testing.T) {
	body := makeBody(100000)
	srv := slowRangeServer(body, 512, 20*time.Millisecond)
	defer srv.Close()

	m := newTestManager(t)
	id, err := m.Add(srv.URL + "/f.bin")
	if err != nil {
		t.Fatal(err)
	}
	waitFor(t, func() bool { st, _ := statusOf(m, id); return st.Downloaded > 0 }, 3*time.Second)

	m.Pause(id)
	waitFor(t, func() bool { return stateOf(m, id) == StatePaused }, 3*time.Second)

	st, _ := statusOf(m, id)
	if st.Downloaded >= st.Size {
		t.Fatal("download completed before pause took effect")
	}
	before := st.Downloaded
	time.Sleep(250 * time.Millisecond)
	st2, _ := statusOf(m, id)
	if st2.Downloaded != before {
		t.Fatalf("bytes grew after pause: %d -> %d", before, st2.Downloaded)
	}
}
