package local

import (
	"strconv"
)

// PrettyPrintFileSizes takes file sizes in int and returns a human readable size e.g. "140mb" as a string.
func PrettyPrintFileSizes(filesize int64) string {

	switch {

	case filesize < 1000:
		return strconv.Itoa(int(filesize)) + " bytes"

	case filesize < 1_000_000:
		filesize /= 1_000
		return strconv.Itoa(int(filesize)) + " kb"

	case filesize < 1_000_000_000:
		filesize /= 1_000_000
		return strconv.Itoa(int(filesize)) + " mb"

	case filesize < 1_000_000_000_000:
		filesize /= 1_000_000_000
		return strconv.Itoa(int(filesize)) + " gb"
	}

	return ""
}
