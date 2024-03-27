package downloader

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
	"wget/internal/types"
)

const (
	buffSize = 1024
	OneKB    = 1024
	OneMB    = 1024 * 1024
)

var (
	ErrInvalidRateLimit = fmt.Errorf("error: invalid rate limit format (use 200K or 2M)")
)

func parseRateLimit(rateLimit string) (float64, error) {
	if rateLimit == "" {
		return 0, nil
	}
	rateRgx := regexp.MustCompile(`^(\d+)([K|M])$`)

	match := rateRgx.FindStringSubmatch(rateLimit)
	if match == nil {
		return 0, ErrInvalidRateLimit
	}

	rate, err := strconv.ParseFloat(match[1], 64)
	if err != nil {
		return 0, ErrInvalidRateLimit
	}

	switch match[2] {
	case "K":
		return rate * OneKB, nil
	case "M":
		return rate * OneMB, nil
	default:
		return 0, ErrInvalidRateLimit
	}
}

func getNameFromURL(url string) string {
	return url[strings.LastIndex(url, "/")+1:]
}

func printProgressBar(currentSize int64, totalSize int64, currentSpeed float64) {
	blocksCount := 50
	activeBlocksCount := int(currentSize * int64(blocksCount) / totalSize)
	activeBlocks := strings.Repeat("=", activeBlocksCount)
	emptyBlocks := strings.Repeat(" ", blocksCount-activeBlocksCount)
	fmt.Printf("\r[%s%s] %.2f%% %.2fM/s", activeBlocks, emptyBlocks, float64(currentSize*100)/float64(totalSize), currentSpeed/OneMB)
}

func createFile(savePath string) (*os.File, error) {
	return os.OpenFile(savePath, os.O_CREATE|os.O_WRONLY, 0644)
}

func download(resp *http.Response, dlInfo chan<- types.DownloadInfo, dlErr chan<- error, file *os.File, rateLimit float64) {
	var receivedBytes int64
	startTime := time.Now()

	buf := make([]byte, buffSize)
	for {
		n, err := resp.Body.Read(buf)
		if err != nil && err != io.EOF {
			dlErr <- fmt.Errorf("error while reading file: %s", err.Error())
			break
		}
		if n == 0 {
			break
		}
		_, err = file.Write(buf[:n])
		if err != nil {
			dlErr <- fmt.Errorf("error while writing to file: %s", err.Error())
			break
		}
		receivedBytes += int64(n)
		speed := float64(receivedBytes) / time.Since(startTime).Seconds()
		if rateLimit != 0 {
			if speed > rateLimit {
				time.Sleep(100 * time.Millisecond)
			}
		}
		if dlInfo != nil {
			data := types.DownloadInfo{
				CurrentSpeed: speed,
				CurrentSize:  receivedBytes,
				TotalSize:    resp.ContentLength,
			}
			dlInfo <- data
		}
	}
}

func averageSpeed(dlInfo <-chan types.DownloadInfo, errCh chan<- error, writer io.Writer) {
	var avgSpeed float64
	var totalSpeed float64
	var total float64

	for info := range dlInfo {
		total++
		totalSpeed += info.CurrentSpeed
	}
	avgSpeed = totalSpeed / total
	_, err := writer.Write([]byte(fmt.Sprintf("Average speed: %.02f\n", avgSpeed)))
	if err != nil {
		errCh <- err
	}
}

func pathCheck(path string) bool {
	if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
		return false
	}
	return true
}

func displayInfo(dlInfo <-chan types.DownloadInfo, dlErr <-chan error, wg *sync.WaitGroup) {
	defer wg.Done()
	for {
		select {
		case info, ok := <-dlInfo:
			if !ok {
				return
			}
			printProgressBar(info.CurrentSize, info.TotalSize, info.CurrentSpeed)
		case err := <-dlErr:
			if err != nil {
				fmt.Print(err.Error())
				return
			}
		}
	}
}

func sendRequest(url string, path string, writer io.Writer) (*http.Response, error) {
	_, err := writer.Write([]byte(fmt.Sprintf("started at %s\n", time.Now().Format("2013-11-14 03:42:06"))))
	_, err = writer.Write([]byte(fmt.Sprintf("sending request on %s, awainting response...\n", url)))
	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("getting response error: %w", err.Error())
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("response from %s: %d, canceled", url, resp.StatusCode)
	}
	_, err = writer.Write([]byte(fmt.Sprintf("Status OK\n")))
	_, err = writer.Write([]byte(fmt.Sprintf("Content size: ~%dKb\n", resp.ContentLength/1024)))
	_, err = writer.Write([]byte(fmt.Sprintf("Content saved to path: %s\n", path)))
	return resp, nil
}

func parseUrlsFromFiles(filePath string) ([]string, error) {
	file, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("urls file parsing error: %w", err)
	}

	fileInStr := strings.TrimSpace(string(file))

	urls := strings.Split(fileInStr, "\n")
	if len(urls) == 0 {
		return nil, fmt.Errorf("no urls in file")
	}
	return urls, nil
}
