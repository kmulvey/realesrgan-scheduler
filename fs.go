package main

import (
	"regexp"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/kmulvey/path"
)

func watchDir(dir string, originalImages chan string) error {

	var suffixRegex = regexp.MustCompile(".*.jpg$|.*.jpeg$|.*.png$|.*.webp$")
	var files = make(chan path.Entry)
	var shutdown = make(chan struct{})
	var watchError = make(chan error)

	// file chan reader
	go func() {
		for file := range files {
			originalImages <- file.AbsolutePath
			log.WithFields(log.Fields{
				"original": file.AbsolutePath,
			}).Info("queued")
		}
	}()

	// file chan writer
	go func() {
		watchError <- path.WatchDirWithFilter(dir, suffixRegex, time.Second*5, files, shutdown)
		close(watchError)
	}()

	return <-watchError
}
