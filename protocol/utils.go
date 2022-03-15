package protocol

import (
	"fmt"
	"time"
)

// GetSwapInfoFilepath returns an info file path with the current timestamp.
func GetSwapInfoFilepath(basepath string) string {
	t := time.Now().Format("2006-Jan-2-15:04:05")
	path := fmt.Sprintf("%s/info-%s.txt", basepath, t)
	return path
}
