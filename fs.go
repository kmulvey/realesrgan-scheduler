package main

import (
	"regexp"

	"github.com/fsnotify/fsnotify"
	"github.com/kmulvey/path"
)

func getExistingFiles(dir string, originalImages chan path.WatchEvent) error {

	// get any files that may already be in the dir because they will not trigger events
	var files, err = path.ListFiles(dir, path.NewRegexFilesFilter(regexp.MustCompile(".*.jpg$|.*.jpeg$|.*.png$|.*.webp$")))
	if err != nil {
		return err
	}
	for _, f := range files {
		originalImages <- path.WatchEvent{Entry: f, Op: fsnotify.Create}
	}

	return nil
}
