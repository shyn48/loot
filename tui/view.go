package tui

import (
	"fmt"
	"image/color"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"simple-gui/core"
	"simple-gui/helper"
	"simple-gui/theme"
)

func hex(c color.RGBA) string {
	return fmt.Sprintf("#%02x%02x%02x", c.R, c.G, c.B)
}

var (
	titleStyle    = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(hex(theme.Accent)))
	mutedStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color(hex(theme.TextMuted)))
	textStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color(hex(theme.Text)))
	accentStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color(hex(theme.Accent)))
	successStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color(hex(theme.Success)))
	dangerStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color(hex(theme.Danger)))
	selectedStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color(hex(theme.Accent)))
)

// bar renders a fixed-width progress bar out of block characters.
func bar(percent float64, width int) string {
	if width <= 0 {
		width = 10
	}
	if percent < 0 {
		percent = 0
	}
	if percent > 100 {
		percent = 100
	}
	filled := int(percent/100*float64(width) + 0.5)
	if filled > width {
		filled = width
	}
	return strings.Repeat("█", filled) + strings.Repeat("░", width-filled)
}

// formatETA renders seconds as m:ss, or an em dash when unknown.
func formatETA(sec int) string {
	if sec < 0 {
		return "—"
	}
	return fmt.Sprintf("%d:%02d", sec/60, sec%60)
}

func truncate(s string, n int) string {
	r := []rune(s)
	if len(r) <= n {
		return s
	}
	if n <= 1 {
		return string(r[:n])
	}
	return string(r[:n-1]) + "…"
}

func (m Model) View() string {
	var b strings.Builder

	active := 0
	var totalSpeed float64
	for _, r := range m.rows {
		if r.State == core.StateDownloading {
			active++
			totalSpeed += r.Speed
		}
	}
	b.WriteString(titleStyle.Render("godownloader"))
	b.WriteString("   ")
	b.WriteString(mutedStyle.Render(fmt.Sprintf("%d active · %s/s", active, helper.HumanBytes(int(totalSpeed)))))
	b.WriteString("\n\n")

	if m.errMsg != "" {
		b.WriteString(dangerStyle.Render("⚠ " + m.errMsg))
		b.WriteString("\n\n")
	}

	if len(m.rows) == 0 {
		b.WriteString(mutedStyle.Render("No downloads yet — press 'a' to add one."))
		b.WriteString("\n\n")
	} else {
		b.WriteString(mutedStyle.Render("  "))
		b.WriteString(mutedStyle.Width(26).Render("NAME"))
		b.WriteString(mutedStyle.Width(22).Render("PROGRESS"))
		b.WriteString(mutedStyle.Width(11).Align(lipgloss.Right).Render("SIZE"))
		b.WriteString(mutedStyle.Width(13).Align(lipgloss.Right).Render("SPEED"))
		b.WriteString(mutedStyle.Width(8).Align(lipgloss.Right).Render("ETA"))
		b.WriteString("\n")
		for i, r := range m.rows {
			b.WriteString(m.renderRow(i, r))
			b.WriteString("\n")
		}
		b.WriteString("\n")
	}

	if m.adding {
		b.WriteString(textStyle.Render("Add URL: "))
		b.WriteString(m.input.View())
		b.WriteString("\n\n")
	}

	b.WriteString(m.footer())
	return b.String()
}

func (m Model) renderRow(i int, r core.JobStatus) string {
	cursor := "  "
	nameStyle := textStyle
	if i == m.cursor {
		cursor = accentStyle.Render("▸ ")
		nameStyle = selectedStyle
	}
	name := nameStyle.Width(26).Render(truncate(r.Name, 25))
	size := mutedStyle.Width(11).Align(lipgloss.Right).Render(helper.HumanBytes(int(r.Size)))
	eta := mutedStyle.Width(8).Align(lipgloss.Right).Render(etaText(r))
	return cursor + name + m.progressCell(r) + size + m.speedCell(r) + eta
}

func (m Model) progressCell(r core.JobStatus) string {
	const barWidth = 12
	switch r.State {
	case core.StateDone:
		return successStyle.Width(22).Render(bar(100, barWidth) + " done")
	case core.StateFailed:
		return dangerStyle.Width(22).Render("failed")
	case core.StatePaused:
		return mutedStyle.Width(22).Render(bar(r.Percent, barWidth) + fmt.Sprintf(" %3.0f%% ⏸", r.Percent))
	default:
		return accentStyle.Width(22).Render(bar(r.Percent, barWidth) + fmt.Sprintf(" %3.0f%%", r.Percent))
	}
}

func (m Model) speedCell(r core.JobStatus) string {
	var s string
	switch r.State {
	case core.StateDownloading:
		s = helper.HumanBytes(int(r.Speed)) + "/s"
	case core.StatePaused:
		s = "paused"
	default:
		s = "—"
	}
	return mutedStyle.Width(13).Align(lipgloss.Right).Render(s)
}

func etaText(r core.JobStatus) string {
	if r.State != core.StateDownloading {
		return "—"
	}
	return formatETA(r.ETASeconds)
}

func (m Model) footer() string {
	if m.showHelp {
		return mutedStyle.Render("↑/k up · ↓/j down · a add · p pause · r resume · d delete · o open folder · q quit (pauses all)")
	}
	return mutedStyle.Render("a add   p pause   r resume   d delete   o open   ? help   q quit")
}
