package wget

import (
	"wget/internal/types"
)

type DownloadMethod interface {
	Download(flags types.Flags)
}

type Wget struct {
	Downloader DownloadMethod
}

func NewWget() *Wget {
	return &Wget{}
}

func (w *Wget) SetDownloader(method DownloadMethod) {
	w.Downloader = method
}

func (w *Wget) StartDownload(flags types.Flags) {
	w.Downloader.Download(flags)
}
