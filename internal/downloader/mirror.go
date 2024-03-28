package downloader

import (
	"errors"
	"fmt"
	"golang.org/x/net/html"
	"log"
	"net/http"
	"wget/internal/types"
)

type Mirror struct {
	url       string
	path      string
	rateLimit float64
	reject    map[string]struct{}
	exclude   map[string]struct{}
}

func (d *Mirror) Download(flags types.Flags) {
	err := d.processFlags(flags)
	if err != nil {
		log.Fatalf("flags error: %s", err.Error())
	}
	resp, err := http.Get(d.url)

	defer resp.Body.Close()
	z := html.NewTokenizer(resp.Body)
	z.Next()
	data := z.Token().String()
	fmt.Println(data)
}

func (d *Mirror) processFlags(flags types.Flags) error {
	var err error
	if flags.URL == "" {
		return errors.New("url not found. Usage: go run ./cmd/main.go -mirror [OPTIONS] [URL]")
	}
	d.url = flags.URL
	d.path = flags.PathFlag
	if flags.PathFlag != "" {
		if !pathCheck(flags.PathFlag) {
			return fmt.Errorf("path not exists")
		}
	}

	d.rateLimit, err = parseRateLimit(flags.RateFlag)
	if err != nil {
		return err
	}
	d.reject, err = parseArgs(flags.Reject)
	d.exclude, err = parseArgs(flags.Exclude)
	return nil
}
