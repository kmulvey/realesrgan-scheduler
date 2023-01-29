package fs

import (
	"context"
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

// WatchEventToEntry convert path.WatchEvent to path.Entry
func WatchEventToEntry(watchEvents []path.WatchEvent) []path.Entry {

	var entires = make([]path.Entry, len(watchEvents))
	for i, watchEvent := range watchEvents {
		entires[i] = watchEvent.Entry
	}
	return entires
}

// WatchEventToEntry convert path.WatchEvent to path.Entry
func WatchEventChanToEntryChan(watchEvents chan path.WatchEvent) chan path.Entry {

	var entires = make(chan path.Entry)
	for watchEvent := range watchEvents {
		entires <- watchEvent.Entry
	}
	return entires
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

// WatchDir will watch the given dir for new files and will publish the ones not already upsized to
// the given images chan.
func WatchDir(ctx context.Context, inputDir, outputDir string, images chan path.Entry) error {

	var events = make(chan path.WatchEvent)

	go func() {
		for e := range events {
			if !AlreadyUpsized(e.Entry, outputDir) {
				images <- e.Entry
			}
		}
		close(images)
	}()

	if err := path.WatchDir(ctx, inputDir, events, path.NewOpWatchFilter(fsnotify.Create), path.NewRegexWatchFilter(regexp.MustCompile(".*.jpg$|.*.jpeg$|.*.png$|.*.webp$"))); err != nil {
		return fmt.Errorf("error in watchDir: %w", err)
	}
	return nil
}
