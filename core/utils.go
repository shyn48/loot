package core

import (
	"os"
	"simple-gui/helper"
	"strings"
)

func getLinkLastPart(link string) string {
	if link == "" {
		return ""
	}
	if string(link[len(link)-1]) == "/" {
		_, newLink := helper.Pop([]rune(link))
		link = string(newLink)
	}

	splittedLink := strings.Split(link, "/")
	urlLastPart := splittedLink[len(splittedLink)-1]

	return urlLastPart
}

// linkHasExtension reports whether the URL's last path segment already carries a
// file extension (a dot that is neither the first nor the last character), e.g.
// "file.zip" or "a.b.mp4" but not "download" or "dir/".
func linkHasExtension(link string) bool {
	last := getLinkLastPart(link)
	i := strings.LastIndex(last, ".")
	return i > 0 && i < len(last)-1
}

func GetDownloadPath(fileName string) (string, error) {
	dirname, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	return dirname + "/" + DOWNLOAD_PATH + "/" + fileName, nil
}

func GetTempPath() (string, error) {
	dirname, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	return dirname + "/" + TMP_PATH, nil
}
