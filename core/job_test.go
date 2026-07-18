package core

import "testing"

func TestJobStatusPercent(t *testing.T) {
	j := &Job{ID: "1", Filename: "f", Size: 200, AcceptRanges: true, state: StateDownloading}
	st := j.statusWithDownloaded(50)
	if st.Percent < 24.9 || st.Percent > 25.1 {
		t.Fatalf("percent=%f", st.Percent)
	}
	if st.Name != "f" || st.State != StateDownloading || !st.Resumable {
		t.Fatalf("%+v", st)
	}
}

func TestJobNotResumableWithoutRanges(t *testing.T) {
	j := &Job{ID: "2", Filename: "g", Size: 100, AcceptRanges: false, state: StateDownloading}
	if j.statusWithDownloaded(10).Resumable {
		t.Fatal("no Accept-Ranges → not resumable")
	}
}
