package gui

import (
	"fmt"

	g "github.com/AllenDang/giu"

	"simple-gui/core"
	"simple-gui/helper"
)

func progressCell(d core.JobStatus) g.Widget {
	return g.ProgressBar(float32(d.Percent/100)).
		Overlay(fmt.Sprintf("%.0f%%", d.Percent)).
		Size(-1, 0)
}

func statusCell(d core.JobStatus) g.Widget {
	switch d.State {
	case core.StateDownloading:
		return coloredLabel("Downloading", colorAccent)
	case core.StatePaused:
		return coloredLabel("Paused", colorTextMuted)
	case core.StateFailed:
		return coloredLabel("Failed", colorDanger)
	case core.StateDone:
		return coloredLabel("Done", colorSuccess)
	default:
		return coloredLabel(string(d.State), colorTextMuted)
	}
}

func removeButton(id string) g.Widget {
	return g.Style().
		SetColor(g.StyleColorButtonHovered, colorDanger).
		To(g.Button("Remove").OnClick(func() {
			manager.Remove(id)
		}))
}

func actionsCell(d core.JobStatus) g.Widget {
	id := d.ID
	switch d.State {
	case core.StateDownloading:
		return g.Row(
			g.Button("Pause").OnClick(func() { manager.Pause(id) }),
			removeButton(id),
		)
	case core.StatePaused, core.StateFailed:
		widgets := []g.Widget{}
		if d.Resumable || d.State == core.StateFailed {
			widgets = append(widgets, g.Button("Resume").OnClick(func() { manager.Resume(id) }))
		}
		widgets = append(widgets, removeButton(id))
		return g.Row(widgets...)
	default: // done
		return g.Row(
			g.Button("Open folder").OnClick(func() {
				if err := manager.OpenFolder(); err != nil {
					SetBoxError(err.Error())
				}
			}),
			removeButton(id),
		)
	}
}

func buildTableRows() []*g.TableRowWidget {
	downloads := manager.Snapshot()

	rows := make([]*g.TableRowWidget, len(downloads))
	for i, d := range downloads {
		rows[i] = g.TableRow(
			g.Label(d.Name),
			progressCell(d),
			g.Label(helper.HumanBytes(int(d.Size))),
			statusCell(d),
			actionsCell(d),
		)
	}
	return rows
}

func downloadsTable() g.Widget {
	header := g.TableRow(
		coloredLabel("NAME", colorTextMuted),
		coloredLabel("PROGRESS", colorTextMuted),
		coloredLabel("SIZE", colorTextMuted),
		coloredLabel("STATUS", colorTextMuted),
		coloredLabel("ACTIONS", colorTextMuted),
	)

	rows := append([]*g.TableRowWidget{header}, buildTableRows()...)

	return g.Table().
		Flags(g.TableFlagsRowBg|g.TableFlagsBordersInnerH|g.TableFlagsResizable|g.TableFlagsScrollY).
		Columns(
			g.TableColumn("name").Flags(g.TableColumnFlagsWidthStretch),
			g.TableColumn("progress").Flags(g.TableColumnFlagsWidthFixed).InnerWidthOrWeight(160),
			g.TableColumn("size").Flags(g.TableColumnFlagsWidthFixed).InnerWidthOrWeight(90),
			g.TableColumn("status").Flags(g.TableColumnFlagsWidthFixed).InnerWidthOrWeight(110),
			g.TableColumn("actions").Flags(g.TableColumnFlagsWidthFixed).InnerWidthOrWeight(220),
		).
		Rows(rows...)
}

func emptyState() g.Widget {
	return g.Align(g.AlignCenter).To(
		g.Layout{
			g.Dummy(0, 60),
			coloredLabel("No downloads yet", colorText),
			g.Dummy(0, 4),
			coloredLabel("Click \"New download\" to grab a file from a URL.", colorTextMuted),
		},
	)
}

func header() g.Widget {
	return g.Layout{
		g.Row(
			g.Style().SetFontSize(22).To(coloredLabel("Loot", colorText)),
		),
		g.Style().SetFontSize(13).To(coloredLabel("Fast, simple downloads.", colorTextMuted)),
		g.Dummy(0, 6),
		primaryButton("+  New download", func() {
			showInputWindow()
		}),
		g.Dummy(0, 6),
		g.Separator(),
		g.Dummy(0, 6),
	}
}

func showErrors() {
	if GetBoxError() != "" {
		g.Msgbox("Error", GetBoxError()).ResultCallback(func(_ g.DialogResult) {
			SetBoxError("")
		})
	}
}

func showMainWindow(mainWindow *g.WindowWidget) {
	var content g.Widget = downloadsTable()
	if len(manager.Snapshot()) == 0 {
		content = emptyState()
	}

	mainWindow.Layout(
		baseStyle().To(
			header(),
			content,
			g.PrepareMsgbox(),
		),
	)
}
