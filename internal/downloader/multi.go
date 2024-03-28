package downloader

import (
	"fmt"
	"io"
	"log"
	"os"
	"path"
	"sync"
	"wget/internal/types"
)

type Multi struct {
	urls      []string
	path      string
	speedRate float64
}

func (d *Multi) Download(flags types.Flags) {
	err := d.processFlags(flags)
	if err != nil {
		log.Fatal(err.Error())
	}

	writer := os.Stdout

	var wg sync.WaitGroup
	var mu sync.Mutex

	errCh := make(chan error)

	for _, url := range d.urls {
		wg.Add(1)
		go d.asyncDownload(&wg, &mu, writer, url, errCh)
	}
	go func() {
		wg.Wait()
		close(errCh)
	}()
	for err = range errCh {
		log.Fatal(err)
	}
}

func (d *Multi) asyncDownload(wg *sync.WaitGroup, mu *sync.Mutex, writer io.Writer, url string, errCh chan<- error) {
	defer wg.Done()
	mu.Lock()
	resp, err := sendRequest(url, d.path, writer)
	if err != nil {
		errCh <- err
	}
	mu.Unlock()

	name := getNameFromURL(url)
	file, err := createFile(path.Join(d.path, name))
	defer file.Close()
	if err != nil {
		errCh <- err
	}

	download(resp, nil, errCh, file, d.speedRate)
	mu.Lock()
	_, err = writer.Write([]byte(fmt.Sprintf("Downloaded: %s\n", name)))
	mu.Unlock()
	if err != nil {
		errCh <- err
	}

}

func (d *Multi) processFlags(flags types.Flags) error {
	var err error
	d.urls, err = parseUrlsFromFiles(flags.MultiFlag)

	d.path = flags.PathFlag
	if flags.PathFlag != "" {
		if !pathCheck(flags.PathFlag) {
			return fmt.Errorf("path not exists")
		}
	}

	d.speedRate, err = parseRateLimit(flags.RateFlag)
	if err != nil {
		return err
	}
	return nil
}
