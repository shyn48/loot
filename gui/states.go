package gui

// UI-local state for the giu front-end. All download state now lives in
// core.Manager (see the package-level manager var); these are only view flags,
// touched exclusively on the render/main thread.
var (
	inputWindowShown    = false
	enterUrlError       = ""
	boxError            = ""
	currentDownloadLink = ""
)

func showInputWindow() { inputWindowShown = true }
func hideInputWindow() { inputWindowShown = false }

func GetCurrentDownloadLink() *string { return &currentDownloadLink }
func SetCurrentDownloadLink(value string) {
	currentDownloadLink = value
}

func SetEnterUrlError(value string) { enterUrlError = value }
func GetEnterUrlError() string      { return enterUrlError }

func SetBoxError(value string) { boxError = value }
func GetBoxError() string      { return boxError }
