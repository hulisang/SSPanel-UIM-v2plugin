package speedtest

import (
	"fmt"
	"io"
	"time"
)

const downloadStreamLimit = 6
const maxDownloadDuration = 10 * time.Second
const downloadBufferSize = 4096
const downloadRepeats = 5

var downloadImageSizes = []int{350, 500, 750, 1000, 1500, 2000, 2500, 3000, 3500, 4000}

func (client *client) downloadFile(url string, start time.Time, ret chan int) {
	totalRead := 0
	defer func() {
		ret <- totalRead
	}()

	if time.Since(start) > maxDownloadDuration {
		if !client.opts.Quiet {
			newErrorf("[%s] Download timeout", url).AtWarning().WriteToLog()
		}

		return
	}
	resp, err := client.Get(url)
	if err != nil {
		newErrorf("[%s] Download failed: %v", url, err).AtWarning().WriteToLog()
		return
	}

	defer resp.Body.Close()

	buf := make([]byte, downloadBufferSize)
	for time.Since(start) <= maxDownloadDuration {
		read, err := resp.Body.Read(buf)
		totalRead += read
		if err != nil {
			if err != io.EOF {
				newErrorf("[%s] Download error: %v\n", url, err).AtWarning().WriteToLog()
				return
			}
			break
		}
	}
}

func (server *Server) DownloadSpeed() int {
	client := server.client.(*client)

	starterChan := make(chan int, downloadStreamLimit)
	downloads := downloadRepeats * len(downloadImageSizes)
	resultChan := make(chan int, downloadStreamLimit)
	start := time.Now()

	go func() {
		for _, size := range downloadImageSizes {
			for i := 0; i < downloadRepeats; i++ {
				url := server.RelativeURL(fmt.Sprintf("random%dx%d.jpg", size, size))
				starterChan <- 1
				go func() {
					client.downloadFile(url, start, resultChan)
					<-starterChan
				}()
			}
		}
		close(starterChan)
	}()

	var totalSize int64 = 0

	for i := 0; i < downloads; i++ {
		totalSize += int64(<-resultChan)
	}

	duration := time.Since(start)

	return int(totalSize * int64(time.Second) / int64(duration))
}
