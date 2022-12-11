package main

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/fsnotify/fsnotify"
	"github.com/kmulvey/path"
	log "github.com/sirupsen/logrus"
)

var ImageExtensionRegex = regexp.MustCompile(".*.jpg$|.*.jpeg$|.*.png$|.*.webp$|.*.JPG$|.*.JPEG$|.*.PNG$|.*.WEBP$")

func getExistingFiles(dir string, originalImages chan path.WatchEvent) error {

	// get any files that may already be in the dir because they will not trigger events
	var files, err = path.List(dir, path.NewRegexListFilter(ImageExtensionRegex))
	if err != nil {
		return err
	}
	for _, f := range files {
		originalImages <- path.WatchEvent{Entry: f, Op: fsnotify.Create}
	}

	return nil
}

func watchEventToEntry(watchEvents chan path.WatchEvent) chan path.Entry {
	var output = make(chan path.Entry, 1000)

	go func() {
		for e := range watchEvents {
			output <- e.Entry
		}
		close(output)
	}()

	return output
}

func dedupMiddleware(outputPath string, input chan path.Entry) chan path.Entry {
	var output = make(chan path.Entry, 1000)

	go func() {
		for originalImage := range input {
			// image is the abs path
			var upsizedImagePath = filepath.Base(originalImage.AbsolutePath)
			upsizedImagePath = filepath.Join(outputPath, strings.Replace(upsizedImagePath, filepath.Ext(upsizedImagePath), ".jpg", 1))

			if stat, _ := os.Stat(upsizedImagePath); stat != nil {
				var err = os.Remove(originalImage.AbsolutePath)
				if err != nil {
					log.Errorf("error removing original file after upscale, err: %s", err.Error())
				}
				//log.WithFields(log.Fields{
				//	"queue length": len(input),
				//	"original":     originalImage.AbsolutePath,
				//	"upsized":      upsizedImagePath,
				//}).Info("already exists, skipping and deleting original")
				continue
			}
			output <- originalImage
		}
		close(output)
	}()
	return output
}
