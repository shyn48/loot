package core

import (
	"bytes"
	"context"
	"crypto/sha256"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"
)

func makeBody(n int) []byte {
	b := make([]byte, n)
	for i := range b {
		b[i] = byte(i % 251)
	}
	return b
}

// rangeServer serves body with Range support (via http.ServeContent).
func rangeServer(body []byte) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.ServeContent(w, r, "f.bin", time.Unix(0, 0), bytes.NewReader(body))
	}))
}

// noRangeServer ignores Range headers and always returns the full body (200).
func noRangeServer(body []byte) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Length", itoa(len(body)))
		w.WriteHeader(http.StatusOK)
		w.Write(body)
	}))
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var b []byte
	for n > 0 {
		b = append([]byte{byte('0' + n%10)}, b...)
		n /= 10
	}
	return string(b)
}

func TestSegmentedDownloadMatchesSource(t *testing.T) {
	body := makeBody(5000)
	srv := rangeServer(body)
	defer srv.Close()

	dir := t.TempDir()
	j := &Job{ID: "1", URL: srv.URL, Filename: "f.bin",
		TargetPath: dir + "/f.bin", Size: int64(len(body)), AcceptRanges: true,
		TotalSection: 4, Sections: computeSections(int64(len(body)), 4)}
	if err := j.run(context.Background(), dir); err != nil {
		t.Fatal(err)
	}
	got, _ := os.ReadFile(j.TargetPath)
	if sha256.Sum256(got) != sha256.Sum256(body) {
		t.Fatal("content mismatch")
	}
}

func TestNoRangeFallbackSingleStream(t *testing.T) {
	body := makeBody(3000)
	srv := noRangeServer(body)
	defer srv.Close()

	dir := t.TempDir()
	// AcceptRanges false → single stream straight to target.
	j := &Job{ID: "2", URL: srv.URL, Filename: "g.bin",
		TargetPath: dir + "/g.bin", Size: int64(len(body)), AcceptRanges: false, TotalSection: 20}
	if err := j.run(context.Background(), dir); err != nil {
		t.Fatal(err)
	}
	got, _ := os.ReadFile(j.TargetPath)
	if sha256.Sum256(got) != sha256.Sum256(body) {
		t.Fatal("single-stream content mismatch")
	}
}

func Test206Enforced(t *testing.T) {
	body := makeBody(4000)
	srv := noRangeServer(body) // ignores Range, returns 200
	defer srv.Close()

	dir := t.TempDir()
	// Claims AcceptRanges → segmented, but server returns 200 → must error, not corrupt.
	j := &Job{ID: "3", URL: srv.URL, Filename: "h.bin",
		TargetPath: dir + "/h.bin", Size: int64(len(body)), AcceptRanges: true,
		TotalSection: 4, Sections: computeSections(int64(len(body)), 4)}
	if err := j.run(context.Background(), dir); err == nil {
		t.Fatal("expected error when server ignores Range")
	}
	if _, err := os.Stat(j.TargetPath); err == nil {
		t.Fatal("no target file should be produced on range failure")
	}
}
