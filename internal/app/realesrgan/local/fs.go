package local

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/fsnotify/fsnotify"
	"github.com/kmulvey/path"
	log "github.com/sirupsen/logrus"
)

// ImageExtensionRegex are all the supported image extensions, and the only ones that will be included in file search/globbing.
var ImageExtensionRegex = regexp.MustCompile(".*.jpg$|.*.jpeg$|.*.png$|.*.webp$|.*.JPG$|.*.JPEG$|.*.PNG$|.*.WEBP$")

func GetExistingFiles(dir string, originalImages chan path.WatchEvent) error {

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

// DedupMiddleware is a channel middleware that attempts to remove files from the queue (channel) that have already been upsized.
// This is a "best effort" function in that it is not perfect and deduplication should also be done at upsize time.
// Example: img.jpg has already been upsized and is added to the channel, this func will remove it. However, this func only considers
// images that have already been upsized so if img.jpg is added to the chan multiple times before it is upsized then all entries will
// pass though this func. This func simply helps to
func DedupMiddleware(outputPath string, input chan path.Entry) chan path.Entry {
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
