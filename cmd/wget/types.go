package wget

import (
	"fmt"
)

type DownloadInfo struct {
	CurrentSpeed float64
	CurrentSize  int64
	TotalSize    int64
}

type DownloadStatus struct {
	Status string
}

type Flags struct {
	BgFlag     bool
	RateFlag   string
	PathFlag   string
	NameFlag   string
	MirrorFlag bool
	MultiFlag  bool
	URL        string
}

const (
	buffSize = 1024
	OneKB    = 1024
	OneMB    = 1024 * 1024
)

var (
	ErrInvalidRateLimit = fmt.Errorf("error: invalid rate limit format (use 200K or 2M)")
)
