package wget

import (
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
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
