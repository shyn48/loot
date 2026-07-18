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

func lg(c color.RGBA) lipgloss.Color { return lipgloss.Color(hex(c)) }

var (
	sTitle   = lipgloss.NewStyle().Bold(true).Foreground(lg(theme.Accent))
	sMuted   = lipgloss.NewStyle().Foreground(lg(theme.TextMuted))
	sText    = lipgloss.NewStyle().Foreground(lg(theme.Text))
	sAccent  = lipgloss.NewStyle().Foreground(lg(theme.Accent))
	sSuccess = lipgloss.NewStyle().Foreground(lg(theme.Success))
	sDanger  = lipgloss.NewStyle().Foreground(lg(theme.Danger))
	sWarn    = lipgloss.NewStyle().Foreground(lg(theme.Warning))
	sPurple  = lipgloss.NewStyle().Foreground(lg(theme.Purple))
	sBorder  = lipgloss.NewStyle().Foreground(lg(theme.Border))
)

// Column widths (inside the list panel). Name flexes to fill the remainder.
const (
	wCursor = 2
	wProg   = 26
	wSize   = 10
	wSpeed  = 11
	wEta    = 10
	wStatus = 13
	gaps    = 5 // single spaces between the 6 columns
)

func box(content string, inner int) string {
	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lg(theme.Border)).
		Width(inner).
		Render(content)
}

func (m Model) View() string {
	width := m.w
	if width < 60 {
		width = 100 // fallback before the first WindowSizeMsg
	}
	inner := width - 2

	parts := []string{m.headerPanel(inner)}
	if m.errMsg != "" {
		parts = append(parts, sDanger.Render(" ⚠ "+m.errMsg))
	}
	parts = append(parts, m.listPanel(inner))
	if sel, ok := m.selected(); ok && m.h >= 26 {
		parts = append(parts, m.detailPanel(sel, inner))
	}
	if m.adding {
		parts = append(parts, sText.Render(" Add URL: ")+m.input.View())
	}
	parts = append(parts, m.footerBar(width))
	return lipgloss.JoinVertical(lipgloss.Left, parts...)
}

// ---- header ----

func (m Model) headerPanel(inner int) string {
	s := computeStats(m.rows)
	left := sTitle.Render("godownloader") + sBorder.Render("  │ ") +
		stat("Active", fmt.Sprint(s.active)) + sBorder.Render(" │ ") +
		stat("Total Speed", helper.HumanBytes(int(s.totalSpeed))+"/s") + sBorder.Render(" │ ") +
		stat("Completed", fmt.Sprint(s.completed)) + sBorder.Render(" │ ") +
		stat("Queued", fmt.Sprint(s.queued)) + sBorder.Render(" │ ") +
		stat("Errors", fmt.Sprint(s.errors))
	right := sMuted.Render(m.clockStr)
	return box(padBetween(left, right, inner), inner)
}

func stat(label, val string) string {
	return sMuted.Render(label+": ") + sAccent.Bold(true).Render(val)
}

func padBetween(l, r string, w int) string {
	gap := w - lipgloss.Width(l) - lipgloss.Width(r)
	if gap < 1 {
		gap = 1
	}
	return l + strings.Repeat(" ", gap) + r
}

// ---- list ----

func (m Model) nameWidth(inner int) int {
	w := inner - wCursor - wProg - wSize - wSpeed - wEta - wStatus - gaps
	if w < 10 {
		w = 10
	}
	return w
}

func (m Model) listPanel(inner int) string {
	wName := m.nameWidth(inner)
	var b strings.Builder

	b.WriteString(columnHeader(wName))
	b.WriteString("\n")
	b.WriteString(sBorder.Render(strings.Repeat("╌", inner)))

	if len(m.rows) == 0 {
		b.WriteString("\n\n")
		b.WriteString(sMuted.Render("  No downloads yet — press 'a' to add one."))
		return box(b.String(), inner)
	}

	boundary := activeCount(m.rows)
	for i, r := range m.rows {
		b.WriteString("\n")
		if i == boundary && boundary > 0 && boundary < len(m.rows) {
			b.WriteString(sBorder.Render(strings.Repeat("╌", inner)))
			b.WriteString("\n")
		}
		b.WriteString(m.renderRow(i, r, wName))
	}
	return box(b.String(), inner)
}

func columnHeader(wName int) string {
	col := func(s string, w int, right bool) string {
		st := sMuted.Width(w)
		if right {
			st = st.Align(lipgloss.Right)
		}
		return st.Render(s)
	}
	return strings.Repeat(" ", wCursor) +
		col("Name", wName, false) + " " +
		col("Progress", wProg, false) + " " +
		col("Size", wSize, true) + " " +
		col("Speed", wSpeed, true) + " " +
		col("ETA", wEta, true) + " " +
		col("Status", wStatus, false)
}

func (m Model) renderRow(i int, r core.JobStatus, wName int) string {
	cursor := "  "
	nameStyle := sText
	if i == m.cursor {
		cursor = sAccent.Render("▸ ")
		nameStyle = sAccent.Bold(true)
	}
	name := nameStyle.Width(wName).Render(truncate(r.Name, wName-1))
	size := sMuted.Width(wSize).Align(lipgloss.Right).Render(helper.HumanBytes(int(r.Size)))
	speed := sMuted.Width(wSpeed).Align(lipgloss.Right).Render(speedText(r))
	eta := sMuted.Width(wEta).Align(lipgloss.Right).Render(etaColumn(r))
	return cursor + name + " " + progressColumn(r) + " " + size + " " + speed + " " + eta + " " + statusColumn(r)
}

// ---- cells ----

func stateStyle(s core.JobState) lipgloss.Style {
	switch s {
	case core.StatePaused:
		return sWarn
	case core.StateQueued:
		return sPurple
	case core.StateDone:
		return sSuccess
	case core.StateFailed:
		return sDanger
	default: // downloading / merging
		return sAccent
	}
}

// blockBar renders `width` block segments: filled ones in fStyle, the remainder
// in eStyle, matching the reference's discrete-square look.
func blockBar(percent float64, width int, filled, empty string, fStyle, eStyle lipgloss.Style) string {
	n := int(percent/100*float64(width) + 0.5)
	if n > width {
		n = width
	}
	if n < 0 {
		n = 0
	}
	return fStyle.Render(strings.Repeat(filled, n)) + eStyle.Render(strings.Repeat(empty, width-n))
}

func progressColumn(r core.JobStatus) string {
	const barW = 14
	var bar string
	st := stateStyle(r.State)
	switch r.State {
	case core.StateDone:
		bar = blockBar(100, barW, "■", "■", sSuccess, sSuccess)
	case core.StateQueued:
		bar = blockBar(0, barW, "■", "□", sMuted, sMuted)
	case core.StatePaused:
		bar = blockBar(r.Percent, barW, "■", "·", sWarn, sMuted)
	case core.StateFailed:
		bar = blockBar(r.Percent, barW, "■", "·", sDanger, sMuted)
	default:
		bar = blockBar(r.Percent, barW, "■", "·", sAccent, sMuted)
	}
	pct := st.Render(fmt.Sprintf("%3.0f%%", r.Percent))
	return lipgloss.NewStyle().Width(wProg).Render(bar + " " + pct)
}

func statusColumn(r core.JobStatus) string {
	label := map[core.JobState]string{
		core.StateDownloading: "downloading",
		core.StateMerging:     "merging",
		core.StatePaused:      "paused",
		core.StateQueued:      "queued",
		core.StateDone:        "completed",
		core.StateFailed:      "failed",
	}[r.State]
	if label == "" {
		label = string(r.State)
	}
	return stateStyle(r.State).Width(wStatus).Render(label)
}

func speedText(r core.JobStatus) string {
	if r.State == core.StateDownloading {
		return helper.HumanBytes(int(r.Speed)) + "/s"
	}
	return "—"
}

func etaColumn(r core.JobStatus) string {
	switch r.State {
	case core.StateDownloading:
		return formatHMS(r.ETASeconds)
	case core.StateDone:
		return "00:00:00"
	default:
		return "—"
	}
}

// ---- detail panel (enriched in later tasks) ----

func (m Model) detailPanel(r core.JobStatus, inner int) string {
	var b strings.Builder
	b.WriteString(sMuted.Render("File: ") + sText.Render(r.Name) + "\n")
	b.WriteString(sMuted.Render("Downloaded: ") +
		sText.Render(fmt.Sprintf("%s / %s", helper.HumanBytes(int(r.Downloaded)), helper.HumanBytes(int(r.Size)))) + "\n")
	b.WriteString(sMuted.Render("Progress: ") + progressColumn(r))
	return box(b.String(), inner)
}

// ---- footer ----

func (m Model) footerBar(width int) string {
	keys := []struct{ k, l string }{
		{"a", "add"}, {"p", "pause"}, {"r", "resume"}, {"d", "delete"},
		{"o", "open"}, {"/", "filter"}, {"?", "help"}, {"q", "quit"},
	}
	parts := make([]string, len(keys))
	for i, it := range keys {
		parts[i] = sAccent.Bold(true).Render(it.k) + " " + sMuted.Render(it.l)
	}
	line := " " + strings.Join(parts, sBorder.Render("  │  "))
	rule := sBorder.Render(strings.Repeat("─", width))
	return rule + "\n" + line
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
