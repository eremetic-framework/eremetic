package formatter

import (
	"fmt"
	"time"
)

// FormatTime takes a UnixDate and transforms it into YYYY-mm-dd HH:MM:SS
func FormatTime(unixTime int64) string {
	t := time.Unix(unixTime, 0)

	year, month, day := t.Date()

	return fmt.Sprintf("%d-%d-%d %d:%02d:%d", year, month, day, t.Hour(), t.Minute(), t.Second())
}
