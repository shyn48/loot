package gui

import (
	"simple-gui/core"
	"simple-gui/helper"

	g "github.com/AllenDang/giu"
	"github.com/google/uuid"
)

func startDownloadClick() {
	downloadId := uuid.NewString()

	downloadLink := *GetCurrentDownloadLink()

	if !helper.IsValidUrl(downloadLink) {
		SetEnterUrlError("Entered Link is Not a valid url!")
		return
	}
	hideInputWindow()

	details, err := core.GetFileDetails(downloadLink)
	if err != nil {
		SetBoxError(err.Error())
		return
	}

	addDownload(GuiDownload{
		Id:       downloadId,
		FileName: details.Name,
		Size:     helper.HumanBytes(details.Size),
		State:    DownloadStateDownloading,
	})

	go func(currentLink string, downloadId string, details core.FileDetails) {
		if err := core.StartDownload(currentLink, details); err != nil {
			SetBoxError(err.Error())
			updateDownloadState(downloadId, DownloadStateFailed)
			return
		}
		updateDownloadState(downloadId, DownloadStateDone)
	}(downloadLink, downloadId, details)

	SetEnterUrlError("")
	SetCurrentDownloadLink("")
}

func linkErrorRow() g.Widget {
	if GetEnterUrlError() == "" {
		// Keep vertical space stable whether or not an error is shown.
		return g.Dummy(0, 4)
	}

	return coloredLabel(GetEnterUrlError(), colorDanger)
}

func showLinkWindow(linkWindow *g.WindowWidget, mainWindow *g.WindowWidget) {
	if inputWindowShown {
		if mainWindow.HasFocus() {
			linkWindow.BringToFront()
		}

		linkWindow.Pos(160, 210).IsOpen(&inputWindowShown).Size(480, 210).Layout(
			baseStyle().To(
				g.Style().SetFontSize(17).To(coloredLabel("New download", colorText)),
				g.Dummy(0, 2),
				coloredLabel("Paste the URL of the file you want to download.", colorTextMuted),
				g.Dummy(0, 6),
				g.InputText(GetCurrentDownloadLink()).
					Hint("https://example.com/file.zip").
					Size(-1),
				linkErrorRow(),
				g.Dummy(0, 6),
				g.Row(
					primaryButton("Start Download", startDownloadClick),
					g.Button("Cancel").OnClick(hideInputWindow),
				),
			),
		)
	}
}
