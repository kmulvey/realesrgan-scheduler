package main

import (
	"context"
	"regexp"

	"github.com/fsnotify/fsnotify"
	"github.com/kmulvey/path"
)

func watchDir(dir string, originalImages chan path.WatchEvent) error {
	var ctx = context.Background()
	var suffixRegex = regexp.MustCompile(".*.jpg$|.*.jpeg$|.*.png$|.*.webp$")
	var suffixRegexFilter = path.NewRegexWatchFilter(suffixRegex)
	var opFilter = path.NewOpWatchFilter(fsnotify.Create)

	// get any files that may already be in the dir because they will not trigger events
	var files, err = path.ListFilesWithFilter(dir, suffixRegex)
	if err != nil {
		return err
	}
	for _, f := range files {
		originalImages <- path.WatchEvent{Entry: f, Op: fsnotify.Create}
	}

	return path.WatchDir(ctx, dir, originalImages, opFilter, suffixRegexFilter)
}
