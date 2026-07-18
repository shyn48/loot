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

func TestSparkline(t *testing.T) {
	rows := sparkline([]float64{0, 5, 10}, 3, 2)
	if len(rows) != 2 {
		t.Fatalf("want 2 rows, got %d", len(rows))
	}
	if rows[0] != "  █" {
		t.Fatalf("top row = %q", rows[0])
	}
	if rows[1] != " ██" {
		t.Fatalf("bottom row = %q", rows[1])
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
