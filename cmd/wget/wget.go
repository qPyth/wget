package wget

import (
	"fmt"
	"io"
	"log"
	"log/slog"
	"mime"
	"net/http"
	"os"
	"sync"
	"time"
)

type Wget struct {
	logger       *slog.Logger
	url          string
	inBackground bool
	name         string
	savePath     string
	rateLimit    float64
	isMulti      bool
	isMirror     bool
}

func NewWget(logger *slog.Logger) *Wget {
	return &Wget{logger: logger}
}

func (w *Wget) ProcessFlags(flags Flags) error {
	var err error

	w.rateLimit, err = parseRateLimit(flags.RateFlag)
	if err != nil {
		return err
	}
	if flags.PathFlag != "" {
		if _, err = os.Stat(flags.PathFlag); os.IsNotExist(err) {
			return err
		} else {
			w.savePath = flags.PathFlag
		}
	} else {
		w.savePath = ""
	}
	w.isMulti = flags.MultiFlag
	w.isMirror = flags.MirrorFlag
	w.url = flags.URL
	if w.isMulti {
		w.inBackground = true
	} else {
		w.inBackground = flags.BgFlag
	}

	w.name = flags.NameFlag
	return err
}

func (w *Wget) SetPath(path string) {
	if path != "" {
		if _, err := os.Stat(w.savePath); os.IsNotExist(err) {
			err = os.Mkdir(w.savePath, os.ModePerm)
			if err != nil {
				log.Fatal(err.Error())
			}
		}
	}
	w.savePath = path
}

func (w *Wget) Start() {
	var wg sync.WaitGroup
	dlInfo := make(chan DownloadInfo)
	dlErr := make(chan error)
	dlStatus := make(chan DownloadStatus)

	wg.Add(2)

	go func() {
		defer wg.Done()
		defer close(dlStatus)
		defer close(dlInfo)
		defer close(dlErr)
		w.downloadFile(dlStatus, dlInfo, dlErr)
	}()
	go func() {
		defer wg.Done()
		w.displayInfo(dlStatus, dlInfo, dlErr)

	}()
	wg.Wait()

}

func (w *Wget) downloadFile(dlStatus chan<- DownloadStatus, dlInfo chan<- DownloadInfo, dlErr chan<- error) {

	dlStatus <- DownloadStatus{Status: "Starting download..."}
	dlStatus <- DownloadStatus{Status: fmt.Sprintf("Sending request to %s, awaiting response...", w.url)}
	resp, err := http.Get(w.url)
	if err != nil {
		dlErr <- err
	}
	dlStatus <- DownloadStatus{Status: fmt.Sprintf("Response received, status code: %d", resp.StatusCode)}
	if resp.StatusCode != 200 {
		dlErr <- fmt.Errorf("error: status code %d", resp.StatusCode)
	}
	dlStatus <- DownloadStatus{Status: fmt.Sprintf("File size: %d MB", resp.ContentLength/OneMB)}
	dlStatus <- DownloadStatus{Status: fmt.Sprintf("File will be saved to: %s", w.savePath)}
	defer resp.Body.Close()

	var fileName string
	if w.name != "" {
		ext, err := mime.ExtensionsByType(resp.Header.Get("Content-Type"))
		if err != nil {
			dlErr <- err
		}
		fileName = w.name + ext[0]
	} else {
		fileName = getNameFromURL(w.url)
	}
	fmt.Println(w.savePath + fileName)
	file, err := createFile(w.savePath + fileName)
	defer file.Close()
	w.download(resp, dlInfo, dlErr, file)
	dlStatus <- DownloadStatus{Status: "\nDownload complete"}
}

func (w *Wget) displayInfo(dlStatus <-chan DownloadStatus, dlInfo <-chan DownloadInfo, dlErr <-chan error) {

	for {
		select {
		case status, ok := <-dlStatus:
			if !ok {
				return
			}
			fmt.Printf("%s\n", status.Status)
		case info, ok := <-dlInfo:
			if !ok {
				return
			}
			printProgressBar(info.CurrentSize, info.TotalSize, info.CurrentSpeed)
		case err, ok := <-dlErr:
			if !ok {
				return
			}
			w.logger.Error(err.Error())
			return
		}
	}
}

func (w *Wget) download(resp *http.Response, dlInfo chan<- DownloadInfo, dlErr chan<- error, file *os.File) {
	var receivedBytes int64
	startTime := time.Now()

	buf := make([]byte, buffSize)
	for {
		n, err := resp.Body.Read(buf)
		if err != nil && err != io.EOF {
			dlErr <- fmt.Errorf("error while reading file: %s", err.Error())
		}
		if n == 0 {
			break
		}
		_, err = file.Write(buf[:n])
		if err != nil {
			dlErr <- fmt.Errorf("error while writing to file: %s", err.Error())
		}
		receivedBytes += int64(n)
		speed := float64(receivedBytes) / time.Since(startTime).Seconds()
		if w.rateLimit != 0 {
			if speed > w.rateLimit {
				time.Sleep(100 * time.Millisecond)
			}
		}
		data := DownloadInfo{
			CurrentSpeed: speed,
			CurrentSize:  receivedBytes,
			TotalSize:    resp.ContentLength,
		}
		dlInfo <- data
	}
}
