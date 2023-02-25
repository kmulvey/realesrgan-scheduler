package ignoreregex

import (
	"bufio"
	"os"
	"regexp"
)

func SkipFileToRegexp(filePath string) (*regexp.Regexp, error) {

	readFile, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer readFile.Close()

	var fileScanner = bufio.NewScanner(readFile)

	fileScanner.Split(bufio.ScanLines)

	var regexString string
	for fileScanner.Scan() {
		regexString += fileScanner.Text() + "|"
	}

	regexString = regexString[:len(regexString)-2] // remove trailing |

	return regexp.Compile(regexString)
}
