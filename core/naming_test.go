package core

import "testing"

func TestLinkHasExtension(t *testing.T) {
	cases := map[string]bool{
		"http://x/file.zip":         true,
		"http://x/a.b.mp4":          true,
		"http://x/1Mb.dat":          true,
		"http://x/download":         false,
		"http://x/dir/":             false,
		"http://x/.hidden":          false, // leading dot only
		"http://x/name.":            false, // trailing dot only
		"http://host.com/path/f.7z": true,
	}
	for url, want := range cases {
		if got := linkHasExtension(url); got != want {
			t.Errorf("linkHasExtension(%q) = %v, want %v", url, got, want)
		}
	}
}
