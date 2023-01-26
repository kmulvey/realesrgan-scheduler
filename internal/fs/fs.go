package fs

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/fsnotify/fsnotify"
	"github.com/kmulvey/path"
)

// ImageExtensionRegex are all the supported image extensions, and the only ones that will be included in file search/globbing.
var ImageExtensionRegex = regexp.MustCompile(".*.jpg$|.*.jpeg$|.*.png$|.*.webp$|.*.JPG$|.*.JPEG$|.*.PNG$|.*.WEBP$")

// GetExistingFiles returns a slice of the existing files in the given directory.
func GetExistingFiles(dir string) ([]path.WatchEvent, error) {

	// get any files that may already be in the dir because they will not trigger events
	var files, err = path.List(dir, path.NewRegexListFilter(ImageExtensionRegex))
	if err != nil {
		return nil, err
	}

	var existingFiles = make([]path.WatchEvent, len(files))
	for i, f := range files {
		existingFiles[i] = path.WatchEvent{Entry: f, Op: fsnotify.Create}
	}

	return existingFiles, nil
}

// AlreadyUpsized checks if we already upsized the image.
func AlreadyUpsized(originalImage path.Entry, outputPath string) bool {

	var upsizedImagePath = filepath.Base(originalImage.AbsolutePath)
	upsizedImagePath = filepath.Join(outputPath, strings.Replace(upsizedImagePath, filepath.Ext(upsizedImagePath), ".jpg", 1))

	if _, err := os.Stat(upsizedImagePath); errors.Is(err, os.ErrNotExist) {
		return false
	}

	return true
}

// MakeDir will create a directory if it does not already exist.
func MakeDir(path string) error {
	if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
		err := os.Mkdir(path, os.ModePerm)
		if err != nil {
			return fmt.Errorf("error creating dir: %s, err: %w", path, err)
		}
	}
	return nil
}
