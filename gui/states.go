package gui

import "sync"

var (
	inputWindowShown = false
	enterUrlError    = ""
	boxError         = ""
	currentDownloads = []GuiDownload{}

	// mu guards currentDownloads and boxError, which are written from download
	// goroutines (updateDownloadState / SetBoxError) while the render loop reads
	// them every frame on the main thread.
	mu sync.RWMutex
)

type DownloadState string

const (
	DownloadStateDownloading DownloadState = "DOWNLOADING"
	DownloadStateDone        DownloadState = "DONE"
	DownloadStateFailed      DownloadState = "FAILED"
)

type GuiDownload struct {
	Id       string
	FileName string
	State    DownloadState
	Size     string
}

func removeDownload(id string) {
	mu.Lock()
	defer mu.Unlock()
	for i, download := range currentDownloads {
		if download.Id == id {
			currentDownloads = removeDownloadFromList(currentDownloads, i)
			return
		}
	}
}

func updateDownloadState(id string, newState DownloadState) {
	mu.Lock()
	defer mu.Unlock()
	for i := range currentDownloads {
		if currentDownloads[i].Id == id {
			currentDownloads[i].State = newState
			return
		}
	}
}

func addDownload(download GuiDownload) {
	mu.Lock()
	defer mu.Unlock()
	currentDownloads = append(currentDownloads, download)
}

func getDownloads() []GuiDownload {
	mu.RLock()
	defer mu.RUnlock()
	// Return a copy: callers iterate the result across the frame while download
	// goroutines may mutate the underlying slice.
	out := make([]GuiDownload, len(currentDownloads))
	copy(out, currentDownloads)
	return out
}

func showInputWindow() {
	inputWindowShown = true
}

func hideInputWindow() {
	inputWindowShown = false
}

var (
	currentDownloadLink = ""
)

func GetCurrentDownloadLink() *string {
	return &currentDownloadLink
}

func SetCurrentDownloadLink(value string) {
	currentDownloadLink = value
}

func SetEnterUrlError(value string) {
	enterUrlError = value
}

func GetEnterUrlError() string {
	return enterUrlError
}

func SetBoxError(value string) {
	mu.Lock()
	defer mu.Unlock()
	boxError = value
}

func GetBoxError() string {
	mu.RLock()
	defer mu.RUnlock()
	return boxError
}
