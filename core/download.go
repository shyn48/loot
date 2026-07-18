package core

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
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

type Download struct {
	Url          string
	TargetPath   string
	Filename     string
	TotalSection int
}

// FileDetails is the result of a single HEAD probe: it is resolved once (in the
// GUI, to display the row) and then passed to StartDownload so the download uses
// the exact same filename and never issues a second HEAD request.
type FileDetails struct {
	Name         string
	Size         int
	AcceptRanges bool
}

func (d Download) getFileInfo() (*http.Response, error) {
	r, err := d.getNewRequest("HEAD")
	if err != nil {
		return nil, err
	}

	resp, err := httpClient.Do(r)
	if err != nil {
		return nil, err
	}
	// A HEAD response has no body, but it must still be closed to release the
	// connection. Headers remain readable after Close.
	resp.Body.Close()

	if resp.StatusCode > 299 {
		return nil, fmt.Errorf("status code: %d", resp.StatusCode)
	}

	return resp, nil
}

func (d Download) DownloadSingleThreaded() error {
	r, err := d.getNewRequest("GET")
	if err != nil {
		return err
	}
	resp, err := httpClient.Do(r)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode > 299 {
		return fmt.Errorf("status code: %d", resp.StatusCode)
	}

	f, err := os.OpenFile(d.TargetPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, os.ModePerm)
	if err != nil {
		return err
	}
	defer f.Close()

	if _, err := io.Copy(f, resp.Body); err != nil {
		return err
	}

	return nil
}

func (d Download) Do(size int, acceptRanges bool) error {
	// Fall back to a plain single-stream download when segmented download is
	// impossible or pointless: unknown size, no server range support, or a file
	// too small to split into TotalSection non-empty pieces.
	if size == 0 || !acceptRanges || d.TotalSection <= 1 || size < d.TotalSection {
		return d.DownloadSingleThreaded()
	}

	var sections = make([][2]int, d.TotalSection)
	eachSize := size / d.TotalSection

	for i := range sections {
		if i == 0 {
			// starting byte of first section
			sections[i][0] = 0
		} else {
			// starting byte of next section
			sections[i][0] = sections[i-1][1] + 1
		}

		if i < d.TotalSection-1 {
			// ending byte of other sections
			sections[i][1] = sections[i][0] + eachSize
		} else {
			// ending byte of last section
			sections[i][1] = size - 1
		}
	}

	wg := sync.WaitGroup{}
	// Each goroutine writes only its own index, so this slice needs no lock.
	errs := make([]error, d.TotalSection)

	for i, section := range sections {
		wg.Add(1)
		go func(i int, section [2]int) {
			defer wg.Done()
			errs[i] = d.downloadSection(i, section[0], section[1])
		}(i, section)
	}

	wg.Wait()

	for _, e := range errs {
		if e != nil {
			// Leave no partial temp files behind on failure.
			d.cleanupTempFiles(len(sections))
			return e
		}
	}

	return d.mergeFiles(sections)
}

func (d Download) downloadSection(index int, startByte int, endByte int) error {
	r, err := d.getNewRequest("GET")
	if err != nil {
		return err
	}
	r.Header.Set("Range", fmt.Sprintf("bytes=%d-%d", startByte, endByte))
	resp, err := httpClient.Do(r)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	// If the server ignored the Range header it returns 200 with the FULL body.
	// Writing that into every section would produce a corrupt, oversized merge,
	// so treat anything other than 206 Partial Content as a hard error.
	if resp.StatusCode != http.StatusPartialContent {
		return fmt.Errorf("section %d: server did not honor range request (status %d)", index+1, resp.StatusCode)
	}

	tempPath, err := GetTempPath()
	if err != nil {
		return err
	}

	f, err := os.OpenFile(d.sectionFile(tempPath, index), os.O_CREATE|os.O_WRONLY|os.O_TRUNC, os.ModePerm)
	if err != nil {
		return err
	}
	defer f.Close()

	// Stream straight to disk instead of buffering the whole section in memory.
	if _, err := io.Copy(f, resp.Body); err != nil {
		return err
	}

	return nil
}

func (d Download) getNewRequest(method string) (*http.Request, error) {
	r, err := http.NewRequest(method, d.Url, nil)
	if err != nil {
		return nil, err
	}

	r.Header.Set("User-Agent", "Dl manager v001")

	return r, nil
}

func (d Download) sectionFile(tempPath string, index int) string {
	return fmt.Sprintf("%s/section-%d-%s.tmp", tempPath, index+1, d.Filename)
}

func (d Download) mergeFiles(sections [][2]int) error {
	// O_TRUNC (not O_APPEND) so a stale file at this path is overwritten rather
	// than appended to, which would corrupt the result.
	f, err := os.OpenFile(d.TargetPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, os.ModePerm)
	if err != nil {
		return err
	}
	defer f.Close()

	tempPath, err := GetTempPath()
	if err != nil {
		return err
	}

	for i := range sections {
		sf, err := os.Open(d.sectionFile(tempPath, i))
		if err != nil {
			return err
		}
		_, err = io.Copy(f, sf)
		sf.Close()
		if err != nil {
			return err
		}
	}

	d.cleanupTempFiles(len(sections))

	return nil
}

func (d Download) cleanupTempFiles(count int) {
	tempPath, err := GetTempPath()
	if err != nil {
		return
	}
	for i := 0; i < count; i++ {
		os.Remove(d.sectionFile(tempPath, i))
	}
}

func GetFileDetails(link string) (FileDetails, error) {
	d := Download{
		Url: link,
	}

	resp, err := d.getFileInfo()
	if err != nil {
		return FileDetails{}, err
	}

	var size int
	if resp.Header.Get("Content-Length") != "" {
		size, err = strconv.Atoi(resp.Header.Get("Content-Length"))
		if err != nil {
			return FileDetails{}, err
		}
	}

	acceptRanges := strings.EqualFold(resp.Header.Get("Accept-Ranges"), "bytes")

	// Derive the file extension from Content-Type, guarding against headers that
	// are missing or malformed (the naive Split(...)[1] would panic on those).
	var fileType string
	if ct := resp.Header.Get("Content-Type"); strings.Contains(ct, "/") {
		fileType = strings.SplitN(strings.SplitN(ct, ";", 2)[0], "/", 2)[1]
		fileType = strings.TrimSpace(fileType)
	}

	fileName := getLinkLastPart(link)

	switch {
	case doesLinkIncludeFileName(link, fileType):
		// URL already ends in the right extension; use it as-is.
	case fileType != "":
		fileName += "." + fileType
		if len(getLinkLastPart(link)) > 15 {
			fileName = strconv.Itoa(int(time.Now().UnixMilli())) + "." + fileType
		}
	}

	filePath, err := GetDownloadPath(fileName)
	if err != nil {
		return FileDetails{}, err
	}

	if _, err := os.Stat(filePath); err == nil {
		fileName = strconv.Itoa(int(time.Now().UnixMilli())) + "-" + fileName
	}

	return FileDetails{Name: fileName, Size: size, AcceptRanges: acceptRanges}, nil
}

// StartDownload runs the download for a link whose details were already resolved
// (by the caller's earlier GetFileDetails call), so the on-disk filename matches
// exactly what the UI showed and no second HEAD request is made.
func StartDownload(link string, details FileDetails) error {
	downloadPath, err := GetDownloadPath(details.Name)
	if err != nil {
		return err
	}

	d := Download{
		Url:          link,
		TargetPath:   downloadPath,
		Filename:     details.Name,
		TotalSection: 20,
	}

	return d.Do(details.Size, details.AcceptRanges)
}
