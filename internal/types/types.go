package types

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
	MultiFlag  string
	Reject     string
	Exclude    string
	URL        string
}
