package tui

import (
	"fmt"
	"image/color"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"

	"simple-gui/core"
	"simple-gui/helper"
	"simple-gui/theme"
)

func hex(c color.RGBA) string        { return fmt.Sprintf("#%02x%02x%02x", c.R, c.G, c.B) }
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

// box wraps content in a rounded border with horizontal padding. cw is the
// content width; the rendered box is cw+4 wide (2 border + 2 padding). Note
// lipgloss's Width includes padding, so it is set to cw+2.
func box(content string, cw int) string {
	return lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lg(theme.Border)).
		Padding(0, 1).
		Width(cw + 2).
		Render(content)
}

// cols holds the computed, responsive column widths for the list.
type cols struct {
	name, prog, size, speed, eta, status, barW int
}

func computeCols(cw int) cols {
	const (
		wCursor = 2
		gaps    = 5
		wSize   = 10
		wSpeed  = 10
		wEta    = 9
		wStatus = 12
	)
	avail := cw - wCursor - gaps - wSize - wSpeed - wEta - wStatus
	if avail < 24 {
		avail = 24
	}
	wProg := avail * 42 / 100
	if wProg < 16 {
		wProg = 16
	}
	if wProg > 30 {
		wProg = 30
	}
	wName := avail - wProg
	if wName < 8 {
		wName = 8
		wProg = avail - wName
	}
	return cols{name: wName, prog: wProg, size: wSize, speed: wSpeed, eta: wEta, status: wStatus, barW: wProg - 5}
}

func (m Model) View() string {
	width := m.w
	if width < 60 {
		width = 100
	}
	height := m.h
	if height < 16 {
		height = 30
	}
	if width < 76 {
		return sMuted.Render("Terminal too narrow — widen to at least 76 columns.")
	}
	cw := width - 4 // content width inside border+padding

	sel, hasSel := m.selected()
	showDetail := hasSel && height >= 27

	extra := 0
	if m.errMsg != "" {
		extra++
	}
	if m.adding || m.filtering || m.filter != "" {
		extra++
	}
	detailH := 0
	if showDetail {
		detailH = 7
	}
	listPanelH := height - 3 - 1 - detailH - extra
	if listPanelH < 6 {
		listPanelH = 6
	}
	maxRows := listPanelH - 4
	if maxRows < 1 {
		maxRows = 1
	}

	parts := []string{m.headerPanel(cw)}
	if m.errMsg != "" {
		parts = append(parts, sDanger.Render(" ⚠ "+m.errMsg))
	}
	parts = append(parts, m.listPanel(cw, maxRows))
	if showDetail {
		parts = append(parts, m.detailPanel(sel, cw))
	}
	if m.adding {
		parts = append(parts, sText.Render(" Add URL: ")+m.input.View())
	} else if m.filtering {
		parts = append(parts, sText.Render(" Filter: ")+m.input.View())
	} else if m.filter != "" {
		parts = append(parts, sMuted.Render(" Filter: ")+sAccent.Render(m.filter)+sMuted.Render("  (esc to clear)"))
	}
	if m.showHelp {
		parts = append(parts, sMuted.Render(" ↑/k up · ↓/j down · a add (clipboard/batch) · p pause · r resume · d delete · c clear completed · o open · / filter · q quit"))
	}
	parts = append(parts, m.footerBar(width))
	return lipgloss.JoinVertical(lipgloss.Left, parts...)
}

// ---- header ----

func (m Model) headerPanel(cw int) string {
	s := computeStats(m.rows)
	speed := helper.HumanBytes(int(s.totalSpeed)) + "/s"
	right := sMuted.Render(m.clockStr)

	full := sTitle.Render("Loot") + sBorder.Render("  │ ") +
		stat("Active", fmt.Sprint(s.active)) + sBorder.Render(" │ ") +
		stat("Total Speed", speed) + sBorder.Render(" │ ") +
		stat("Completed", fmt.Sprint(s.completed)) + sBorder.Render(" │ ") +
		stat("Queued", fmt.Sprint(s.queued)) + sBorder.Render(" │ ") +
		stat("Errors", fmt.Sprint(s.errors))
	if lipgloss.Width(full)+lipgloss.Width(right)+1 <= cw {
		return box(padBetween(full, right, cw), cw)
	}

	// Compact header for narrow terminals.
	cs := func(label, val string) string { return sMuted.Render(label) + sAccent.Bold(true).Render(val) }
	compact := sTitle.Render("Loot") + "  " +
		cs("A:", fmt.Sprint(s.active)) + "  " +
		sAccent.Bold(true).Render(speed) + "  " +
		cs("✓", fmt.Sprint(s.completed)) + "  " +
		cs("Q:", fmt.Sprint(s.queued)) + "  " +
		cs("E:", fmt.Sprint(s.errors))
	if lipgloss.Width(compact)+lipgloss.Width(right)+1 <= cw {
		return box(padBetween(compact, right, cw), cw)
	}
	return box(compact, cw)
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

func (m Model) listPanel(cw, maxRows int) string {
	c := computeCols(cw)
	rows := m.visible()
	var b strings.Builder

	b.WriteString(columnHeader(c))
	b.WriteString("\n")
	b.WriteString(sBorder.Render(strings.Repeat("╌", cw)))

	if len(rows) == 0 {
		b.WriteString("\n\n")
		if m.filter != "" {
			b.WriteString(sMuted.Render("  No downloads match \"" + m.filter + "\"."))
		} else {
			b.WriteString(sMuted.Render("  No downloads yet — press 'a' to add one."))
		}
		return box(b.String(), cw)
	}

	start, end := windowRows(len(rows), m.cursor, maxRows)
	boundary := activeCount(rows)
	if start > 0 {
		b.WriteString("\n" + sMuted.Render(fmt.Sprintf("  ↑ %d more", start)))
	}
	for i := start; i < end; i++ {
		b.WriteString("\n")
		if i == boundary && boundary > 0 && boundary < len(rows) {
			b.WriteString(sBorder.Render(strings.Repeat("╌", cw)) + "\n")
		}
		b.WriteString(m.renderRow(i, rows[i], c, cw))
	}
	if end < len(rows) {
		b.WriteString("\n" + sMuted.Render(fmt.Sprintf("  ↓ %d more", len(rows)-end)))
	}
	return box(b.String(), cw)
}

func windowRows(total, cursor, maxRows int) (int, int) {
	if maxRows >= total {
		return 0, total
	}
	start := cursor - maxRows/2
	if start < 0 {
		start = 0
	}
	end := start + maxRows
	if end > total {
		end = total
		start = end - maxRows
	}
	return start, end
}

func columnHeader(c cols) string {
	h := func(s string, w int, right bool) string {
		st := sMuted.Width(w)
		if right {
			st = st.Align(lipgloss.Right)
		}
		return st.Render(s)
	}
	return "  " +
		h("Name", c.name, false) + " " +
		h("Progress", c.prog, false) + " " +
		h("Size", c.size, true) + " " +
		h("Speed", c.speed, true) + " " +
		h("ETA", c.eta, true) + " " +
		h("Status", c.status, false)
}

func (m Model) renderRow(i int, r core.JobStatus, c cols, cw int) string {
	sel := i == m.cursor
	bg := func(s lipgloss.Style) lipgloss.Style {
		if sel {
			return s.Background(lg(theme.Selection))
		}
		return s
	}
	sp := bg(lipgloss.NewStyle()).Render(" ")

	cursor := "  "
	nameStyle := sText
	if sel {
		cursor = "▸ "
		nameStyle = sText.Bold(true)
	}
	row := bg(sAccent).Render(cursor) +
		bg(nameStyle).Width(c.name).Render(truncate(r.Name, c.name-1)) + sp +
		progressColumn(r, c.barW, c.prog, bg) + sp +
		bg(sMuted).Width(c.size).Align(lipgloss.Right).Render(helper.HumanBytes(int(r.Size))) + sp +
		bg(sMuted).Width(c.speed).Align(lipgloss.Right).Render(speedText(r)) + sp +
		bg(sMuted).Width(c.eta).Align(lipgloss.Right).Render(etaColumn(r)) + sp +
		bg(stateStyle(r.State)).Width(c.status).Render(statusLabel(r))
	return row
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
	default:
		return sAccent
	}
}

func statusLabel(r core.JobStatus) string {
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
	return label
}

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

func progressColumn(r core.JobStatus, barW, wProg int, bg func(lipgloss.Style) lipgloss.Style) string {
	st := stateStyle(r.State)
	var bar string
	switch r.State {
	case core.StateDone:
		bar = blockBar(100, barW, "■", "■", bg(sSuccess), bg(sSuccess))
	case core.StateQueued:
		bar = blockBar(0, barW, "■", "□", bg(sMuted), bg(sMuted))
	case core.StatePaused:
		bar = blockBar(r.Percent, barW, "■", "·", bg(sWarn), bg(sMuted))
	case core.StateFailed:
		bar = blockBar(r.Percent, barW, "■", "·", bg(sDanger), bg(sMuted))
	default:
		bar = blockBar(r.Percent, barW, "■", "·", bg(sAccent), bg(sMuted))
	}
	pct := bg(st).Render(fmt.Sprintf("%3.0f%%", r.Percent))
	return bg(lipgloss.NewStyle()).Width(wProg).Render(bar + bg(lipgloss.NewStyle()).Render(" ") + pct)
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

// ---- detail panel ----

func (m Model) detailPanel(r core.JobStatus, cw int) string {
	rightW := 30
	const sepTotal = 6
	leftW := (cw - rightW - sepTotal) / 2
	midW := cw - leftW - rightW - sepTotal
	if leftW < 26 || midW < 14 {
		return box(detailLeft(r, cw), cw)
	}
	noBg := func(s lipgloss.Style) lipgloss.Style { return s }
	row := lipgloss.JoinHorizontal(lipgloss.Top,
		detailLeft(r, leftW),
		sBorder.Render(" │ "),
		m.detailSparkline(r, midW),
		sBorder.Render(" │ "),
		detailRight(r, rightW, noBg),
	)
	return box(row, cw)
}

func detailLeft(r core.JobStatus, w int) string {
	field := func(label, val string, valStyle lipgloss.Style) string {
		return sMuted.Render(label) + valStyle.Render(truncate(val, max(1, w-len(label))))
	}
	lines := []string{
		field("File: ", r.Name, sText),
		field("URL : ", r.URL, sAccent),
		field("Path: ", r.Path, sMuted),
		sMuted.Render("Connections: ") + sText.Render(fmt.Sprint(r.Connections)),
		sMuted.Render("Downloaded: ") + sText.Render(fmt.Sprintf("%s / %s",
			helper.HumanBytes(int(r.Downloaded)), helper.HumanBytes(int(r.Size)))),
	}
	return lipgloss.NewStyle().Width(w).Render(strings.Join(lines, "\n"))
}

func detailRight(r core.JobStatus, w int, bg func(lipgloss.Style) lipgloss.Style) string {
	rule := ""
	if w > 10 {
		rule = sBorder.Render(strings.Repeat("─", w-9))
	}
	started, elapsed := "—", "—"
	if !r.StartedAt.IsZero() {
		started = r.StartedAt.Format("15:04:05")
		elapsed = formatHMS(int(time.Since(r.StartedAt).Seconds()))
	}
	lines := []string{
		sMuted.Render("Progress ") + rule,
		progressColumn(r, w-6, w, bg),
		sMuted.Render("Started: ") + sText.Render(started),
		sMuted.Render("Elapsed: ") + sText.Render(elapsed),
		sMuted.Render("ETA:     ") + sText.Render(etaColumn(r)),
	}
	return lipgloss.NewStyle().Width(w).Render(strings.Join(lines, "\n"))
}

func (m Model) detailSparkline(r core.JobStatus, w int) string {
	const h = 4
	hist := m.speedHist[r.ID]
	var maxV float64
	for _, v := range hist {
		if v > maxV {
			maxV = v
		}
	}
	rows := sparkline(hist, max(1, w-5), h)
	maxMB := maxV / 1e6
	labels := []string{fmt.Sprintf("%3.1f", maxMB), "", fmt.Sprintf("%3.1f", maxMB/2), "0.0"}

	var b strings.Builder
	b.WriteString(sMuted.Render("Speed (MB/s)") + "\n")
	for i, rowStr := range rows {
		lbl := ""
		if i < len(labels) {
			lbl = labels[i]
		}
		b.WriteString(sMuted.Render(fmt.Sprintf("%4s ", lbl)) + sAccent.Render(rowStr) + "\n")
	}
	return lipgloss.NewStyle().Width(w).Render(strings.TrimRight(b.String(), "\n"))
}

// sparkline renders samples as `height` rows of `width` columns using
// partial-block characters for smooth bar tops.
func sparkline(samples []float64, width, height int) []string {
	blocks := []rune(" ▁▂▃▄▅▆▇█")
	if len(samples) > width {
		samples = samples[len(samples)-width:]
	}
	var maxV float64
	for _, v := range samples {
		if v > maxV {
			maxV = v
		}
	}
	if maxV <= 0 {
		maxV = 1
	}
	levels := make([]int, len(samples)) // 0 .. height*8
	for i, v := range samples {
		levels[i] = int(v/maxV*float64(height*8) + 0.5)
	}
	rows := make([]string, height)
	pad := max(0, width-len(samples))
	for r := 0; r < height; r++ {
		base := (height - 1 - r) * 8 // this row covers [base, base+8)
		var sb strings.Builder
		sb.WriteString(strings.Repeat(" ", pad))
		for _, lvl := range levels {
			cell := lvl - base
			switch {
			case cell <= 0:
				sb.WriteRune(' ')
			case cell >= 8:
				sb.WriteRune('█')
			default:
				sb.WriteRune(blocks[cell])
			}
		}
		rows[r] = sb.String()
	}
	return rows
}

// ---- footer ----

func (m Model) footerBar(width int) string {
	base := lipgloss.NewStyle().Background(lg(theme.FooterBg))
	keyStyle := base.Foreground(lg(theme.Accent)).Bold(true)
	lblStyle := base.Foreground(lg(theme.TextMuted))
	sepStyle := base.Foreground(lg(theme.Border))

	keys := []struct{ k, l string }{
		{"a", "add"}, {"p", "pause"}, {"r", "resume"}, {"d", "delete"},
		{"o", "open"}, {"/", "filter"}, {"?", "help"}, {"q", "quit"},
	}
	parts := make([]string, len(keys))
	for i, it := range keys {
		parts[i] = keyStyle.Render(it.k) + lblStyle.Render(" "+it.l)
	}
	line := strings.Join(parts, sepStyle.Render(" │ "))
	return base.Width(width).Padding(0, 1).Render(line)
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
