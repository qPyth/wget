package downloader

import (
	"bytes"
	"errors"
	"fmt"
	"golang.org/x/net/html"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"path"
	"strings"
	"wget/internal/types"
)

type Mirror struct {
	url             string
	path            string
	rateLimit       float64
	reject          map[string]struct{}
	exclude         map[string]struct{}
	downloadedPages map[string]bool
}

func (d *Mirror) Download(flags types.Flags) {
	d.downloadedPages = make(map[string]bool)
	err := d.processFlags(flags)
	if err != nil {
		log.Fatalf("flags error: %s", err.Error())
	}
	fmt.Printf("Downloading %s\n", d.url)
	d.downloadPage(d.url)
	fmt.Println("Downloading complete")
}

func (d *Mirror) downloadPage(url string) {
	absUrl, err := normalizeUrl(d.url, url)
	if absUrl == nil {
		return
	}
	if _, ok := d.downloadedPages[absUrl.String()]; ok {
		return
	}
	resp, err := http.Get(absUrl.String())
	if err != nil {
		log.Printf("http.Get error from url :%s with error %s\n", url, err.Error())
		return
	}
	if resp.StatusCode != http.StatusOK {
		return
	}
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("reading body error: %s, from: %s", err.Error(), absUrl.String())
		return
	}
	err = saveFile(d.path, absUrl, data)
	if err != nil {
		log.Printf("downloading error: %s,file from: %s\n", err.Error(), url)
		return
	}
	d.downloadedPages[absUrl.String()] = true

	doc, err := html.Parse(bytes.NewReader(data))
	if err != nil {
		log.Fatalf("parsing error: %s\n", err.Error())
	}

	d.domTraversing(doc)
}

const (
	tagLink   = "a"
	tagStyle  = "link"
	tagImg    = "img"
	tagScript = "script"
	attrRel   = "rel"
	attrHref  = "href"
	attrSrc   = "src"
)

func (d *Mirror) domTraversing(node *html.Node) {
	if node == nil {
		return
	}
	if node.Type == html.ElementNode {
		if u := nodeParse(node); u != "" {
			d.downloadPage(u)
		}
	}
	for c := node.FirstChild; c != nil; c = c.NextSibling {
		d.domTraversing(c)
	}
}

func (d *Mirror) processFlags(flags types.Flags) error {
	var err error
	if flags.URL == "" {
		return errors.New("url not found. Usage: go run ./cmd/main.go -mirror [OPTIONS] [URL]")
	}
	d.url = flags.URL
	u, err := url.Parse(d.url)
	if err != nil {
		return err
	}
	d.path = path.Join(flags.PathFlag, u.Hostname())
	if d.path == "" {
		d.path = d.url
	}
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

func nodeParse(n *html.Node) string {
	switch n.Data {
	case tagLink, tagStyle:
		return attrParse(n.Attr, attrHref)
	case tagScript, tagImg:
		return attrParse(n.Attr, attrSrc)
	}
	return ""
}

func attrParse(attrs []html.Attribute, key string) string {
	for _, attr := range attrs {
		if attr.Key == key {
			return attr.Val
		}
	}
	return ""
}

func normalizeUrl(baseUrlStr, urlStr string) (*url.URL, error) {
	baseURL, err := url.Parse(baseUrlStr)
	if err != nil {
		return nil, err
	}

	u, err := url.Parse(urlStr)
	if err != nil {
		return nil, err
	}
	u.Fragment = ""

	if u.Host == "" {
		normalizedURL := baseURL.ResolveReference(u)
		return normalizedURL, nil
	}
	if baseURL.Host == u.Host {
		return u, nil
	}
	return nil, nil
}

func saveFile(path string, u *url.URL, data []byte) error {
	pathToFile, err := generateFilePath(path, u)
	if err != nil {
		return err
	}
	file, err := os.Create(pathToFile)
	if err != nil {
		return err
	}
	defer file.Close()
	_, err = file.Write(data)
	if err != nil {
		return err
	}
	return nil
}

func generateFilePath(baseDir string, u *url.URL) (string, error) {
	var filePath string
	if u.Path == "" || strings.HasSuffix(u.Path, "/") {
		filePath = path.Join(u.Path, "index.html")
	} else {
		lastSegment := path.Base(u.Path)
		if strings.Contains(lastSegment, ".") {
			filePath = u.Path
		} else {
			filePath = u.Path + ".html"
		}
	}

	fullPath := path.Join(baseDir, filePath)
	if u.RawQuery != "" {
		fullPath += "?" + u.RawQuery
	}

	dirPath := path.Dir(fullPath)
	if err := os.MkdirAll(dirPath, 0755); err != nil {
		return "", err
	}

	return fullPath, nil
}
