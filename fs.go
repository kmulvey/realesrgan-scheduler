package main

import (
	"regexp"
	"time"

	"github.com/kmulvey/path"
)

func watchDir(dir string, originalImages chan path.Entry) error {
	var suffixRegex = regexp.MustCompile(".*.jpg$|.*.jpeg$|.*.png$|.*.webp$")
	var shutdown = make(chan struct{})
	return path.WatchDirWithFilter(dir, suffixRegex, time.Second*5, originalImages, shutdown)
}
