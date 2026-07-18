package tui

import (
	"strings"
	"testing"
)

func TestBar(t *testing.T) {
	if got := bar(0, 10); got != strings.Repeat("░", 10) {
		t.Fatalf("bar(0) = %q", got)
	}
	if got := bar(100, 10); got != strings.Repeat("█", 10) {
		t.Fatalf("bar(100) = %q", got)
	}
	if got := bar(50, 10); got != strings.Repeat("█", 5)+strings.Repeat("░", 5) {
		t.Fatalf("bar(50) = %q", got)
	}
}

func TestFormatETA(t *testing.T) {
	if formatETA(-1) != "—" {
		t.Fatal("eta -1 should be em dash")
	}
	if got := formatETA(72); got != "1:12" {
		t.Fatalf("formatETA(72) = %q", got)
	}
	if got := formatETA(5); got != "0:05" {
		t.Fatalf("formatETA(5) = %q", got)
	}
}
