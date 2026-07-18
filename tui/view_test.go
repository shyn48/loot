package tui

import (
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
)

func TestBlockBar(t *testing.T) {
	plain := lipgloss.NewStyle()
	if got := blockBar(0, 10, "■", "·", plain, plain); got != strings.Repeat("·", 10) {
		t.Fatalf("blockBar(0) = %q", got)
	}
	if got := blockBar(100, 10, "■", "·", plain, plain); got != strings.Repeat("■", 10) {
		t.Fatalf("blockBar(100) = %q", got)
	}
	if got := blockBar(50, 10, "■", "·", plain, plain); got != strings.Repeat("■", 5)+strings.Repeat("·", 5) {
		t.Fatalf("blockBar(50) = %q", got)
	}
}

func TestTruncate(t *testing.T) {
	if truncate("short", 10) != "short" {
		t.Fatal("no truncation when it fits")
	}
	if got := truncate("abcdefghij", 5); got != "abcd…" {
		t.Fatalf("truncate = %q", got)
	}
}
