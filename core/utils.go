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

func doesLinkIncludeFileName(link string, contentType string) bool {
	linkLastPart := getLinkLastPart(link)
	splittedLastPart := strings.Split(linkLastPart, ".")

	return len(splittedLastPart) >= 2 && splittedLastPart[len(splittedLastPart)-1] == contentType
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
