package gui

import (
	"fmt"
	"os/exec"
	"simple-gui/core"
	"time"

	g "github.com/AllenDang/giu"
)

// downloadingDots returns an animated "..." suffix so an in-flight download
// reads as active without inventing a fake percentage (real per-download
// progress is not reported by core).
func downloadingDots() string {
	n := (time.Now().UnixMilli() / 400) % 4
	dots := ""
	for i := int64(0); i < n; i++ {
		dots += "."
	}
	return dots
}

func openDownloadFolder() {
	downloadPath, err := core.GetDownloadPath("")
	if err != nil {
		SetBoxError(err.Error())
		return
	}

	cmd := exec.Command("open", downloadPath)
	if err := cmd.Run(); err != nil {
		SetBoxError(err.Error())
	}
}

func statusCell(download GuiDownload) g.Widget {
	if download.State == DownloadStateDownloading {
		return coloredLabel(fmt.Sprintf("Downloading%s", downloadingDots()), colorAccent)
	}

	return coloredLabel("Done", colorSuccess)
}

func actionsCell(download GuiDownload) g.Widget {
	if download.State == DownloadStateDownloading {
		return coloredLabel("in progress", colorTextMuted)
	}

	currentID := download.Id

	return g.Row(
		g.Button("Open folder").OnClick(openDownloadFolder),
		g.Style().
			SetColor(g.StyleColorButtonHovered, colorDanger).
			To(g.Button("Remove").OnClick(func() {
				removeDownload(currentID)
			})),
	)
}

func buildTableRows() []*g.TableRowWidget {
	downloads := getDownloads()

	rows := make([]*g.TableRowWidget, len(downloads))

	for i, download := range downloads {
		rows[i] = g.TableRow(
			g.Label(download.FileName),
			g.Label(download.Size),
			statusCell(download),
			actionsCell(download),
		)
	}

	return rows
}

func downloadsTable() g.Widget {
	header := g.TableRow(
		coloredLabel("NAME", colorTextMuted),
		coloredLabel("SIZE", colorTextMuted),
		coloredLabel("STATUS", colorTextMuted),
		coloredLabel("ACTIONS", colorTextMuted),
	)

	rows := append([]*g.TableRowWidget{header}, buildTableRows()...)

	return g.Table().
		Flags(g.TableFlagsRowBg|g.TableFlagsBordersInnerH|g.TableFlagsResizable|g.TableFlagsScrollY).
		Columns(
			g.TableColumn("name").Flags(g.TableColumnFlagsWidthStretch),
			g.TableColumn("size").Flags(g.TableColumnFlagsWidthFixed).InnerWidthOrWeight(110),
			g.TableColumn("status").Flags(g.TableColumnFlagsWidthFixed).InnerWidthOrWeight(140),
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
			g.Style().SetFontSize(22).To(coloredLabel("Shyn Download Manager", colorText)),
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
	if len(getDownloads()) == 0 {
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
