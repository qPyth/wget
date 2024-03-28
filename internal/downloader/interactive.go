package downloader

import (
	"fmt"
	"log"
	"os"
	"path"
	"sync"
	"wget/internal/types"
)

type Interactive struct {
	url        string
	path       string
	name       string
	speedRate  float64
	background bool
}

func (d *Interactive) Download(flags types.Flags) {
	err := d.processFlags(flags)
	if err != nil {
		log.Fatalf("flags error: %s", err.Error())
	}

	writer := os.Stdout
	dlInfo := make(chan types.DownloadInfo)
	dlErr := make(chan error)

	resp, err := sendRequest(d.url, d.path, writer)
	if err != nil {
		log.Fatal(err.Error())
	}

	file, err := createFile(path.Join(d.path, d.name))
	if err != nil {
		log.Fatal(err.Error())
	}
	defer file.Close()

	var wg sync.WaitGroup
	wg.Add(2)

	if d.background {
		go averageSpeed(dlInfo, dlErr, writer, &wg)
	} else {
		go displayInfo(dlInfo, dlErr, &wg)
	}
	go func(wg *sync.WaitGroup) {
		wg.Done()
		defer close(dlErr)
		defer close(dlInfo)
		download(resp, dlInfo, dlErr, file, d.speedRate)
	}(&wg)

	wg.Wait()
	fmt.Println("\nDownloaded complete")
}

func (d *Interactive) processFlags(flags types.Flags) error {
	var err error
	d.url = flags.URL
	d.background = flags.BgFlag
	d.name = flags.NameFlag
	if flags.NameFlag == "" {
		d.name = getNameFromURL(flags.URL)
	}
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
