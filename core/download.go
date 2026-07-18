package core

import (
	"fmt"
	"net"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// httpClient uses connection-level timeouts rather than an overall
// Client.Timeout: a stalled connection or slow server fails fast, but a large
// (and therefore legitimately slow) response body is never cut off mid-download.
var httpClient = &http.Client{
	Transport: &http.Transport{
		DialContext:           (&net.Dialer{Timeout: 15 * time.Second}).DialContext,
		TLSHandshakeTimeout:   15 * time.Second,
		ResponseHeaderTimeout: 30 * time.Second,
	},
}

// FileDetails is the result of a single HEAD probe: resolved once when a
// download is added, then reused so no second HEAD is issued and the displayed
// name matches the saved file.
type FileDetails struct {
	Name         string
	Size         int64
	AcceptRanges bool
}

func newRequest(url, method string) (*http.Request, error) {
	r, err := http.NewRequest(method, url, nil)
	if err != nil {
		return nil, err
	}
	r.Header.Set("User-Agent", "Dl manager v001")
	return r, nil
}

func getFileInfo(url string) (*http.Response, error) {
	r, err := newRequest(url, "HEAD")
	if err != nil {
		return nil, err
	}
	resp, err := httpClient.Do(r)
	if err != nil {
		return nil, err
	}
	// HEAD has no body, but it must still be closed to release the connection.
	// Headers remain readable after Close.
	resp.Body.Close()
	if resp.StatusCode > 299 {
		return nil, fmt.Errorf("status code: %d", resp.StatusCode)
	}
	return resp, nil
}

func GetFileDetails(link string) (FileDetails, error) {
	resp, err := getFileInfo(link)
	if err != nil {
		return FileDetails{}, err
	}

	var size int64
	if resp.Header.Get("Content-Length") != "" {
		size, err = strconv.ParseInt(resp.Header.Get("Content-Length"), 10, 64)
		if err != nil {
			return FileDetails{}, err
		}
	}

	acceptRanges := strings.EqualFold(resp.Header.Get("Accept-Ranges"), "bytes")

	// Derive the file extension from Content-Type, guarding against headers that
	// are missing or malformed (a naive Split(...)[1] would panic on those).
	var fileType string
	if ct := resp.Header.Get("Content-Type"); strings.Contains(ct, "/") {
		fileType = strings.SplitN(strings.SplitN(ct, ";", 2)[0], "/", 2)[1]
		fileType = strings.TrimSpace(fileType)
	}

	fileName := getLinkLastPart(link)

	switch {
	case linkHasExtension(link):
		// URL already ends in a filename with an extension; use it as-is.
	case fileType != "":
		fileName += "." + fileType
		if len(getLinkLastPart(link)) > 15 {
			fileName = strconv.Itoa(int(time.Now().UnixMilli())) + "." + fileType
		}
	}

	return FileDetails{Name: fileName, Size: size, AcceptRanges: acceptRanges}, nil
}
