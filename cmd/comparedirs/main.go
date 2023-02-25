package main

import (
	"flag"
	"fmt"
	"path/filepath"
	"strconv"

	log "github.com/sirupsen/logrus"

	"github.com/kmulvey/path"
	"github.com/kmulvey/realesrgan-scheduler/internal/pkg/ignoreregex"
)

func main() {
	var originalImages path.Path
	var upscaledImages path.Path
	var skipFile string

	flag.Var(&originalImages, "originals-dir", "")
	flag.Var(&upscaledImages, "upscaled-dir", "")
	flag.StringVar(&skipFile, "skip-file", "", "file with directories to skip, one per line")
	flag.Parse()

	for _, dir := range path.FilterEntities(upscaledImages.Files, path.NewDirEntitiesFilter()) {

		var upsizedBase = filepath.Base(dir.AbsolutePath)
		var leafDir = filepath.Join(originalImages.ComputedPath.AbsolutePath, upsizedBase)

		var re, err = ignoreregex.SkipFileToRegexp(skipFile)
		if err != nil {
			log.Fatalf("error creating regex to skip dirs: %s", err)
		}

		if re.MatchString(leafDir) {
			log.Infof("skipping %s ...", dir)
			continue
		}

		var baseDir = filepath.Base(dir.AbsolutePath)
		originalfiles, err := path.List(filepath.Join(originalImages.ComputedPath.AbsolutePath, baseDir), path.NewFileListFilter())
		if err != nil {
			log.Fatal(err)
		}

		upsizedfiles, err := path.List(dir.AbsolutePath, path.NewFileListFilter())
		if err != nil {
			log.Fatal(err)
		}
		processDir(dir, originalfiles, upsizedfiles)
	}
}

func processDir(dir path.Entry, originalImages, upscaledImages []path.Entry) {
	var originalsMap = make(map[string]path.Entry)
	for _, image := range originalImages {
		originalsMap[filepath.Base(image.String())] = image
	}
	for _, upscaledImage := range upscaledImages {
		delete(originalsMap, filepath.Base(upscaledImage.AbsolutePath))
	}

	if len(originalsMap) > 0 {
		fmt.Printf("\n %s \n", dir.AbsolutePath)
		fmt.Printf("total: %d \n", len(originalsMap))

		if len(originalsMap) < 50 {
			for originalImageName, originalImage := range originalsMap {
				if originalImage.FileInfo.Size() < 1_000_000 {
					fmt.Printf("%s %s \n", prettyPrintFileSizes(originalImage.FileInfo.Size()), originalImageName)
				}
			}
		}
	}
}

func prettyPrintFileSizes(filesize int64) string {
	if filesize < 1_000 {
		return strconv.Itoa(int(filesize)) + " bytes"
	} else if filesize < 1_000_000 {
		filesize /= 1_000
		return strconv.Itoa(int(filesize)) + " kb"
	} else if filesize < 1_000_000_000 {
		filesize /= 1_000_000
		return strconv.Itoa(int(filesize)) + " mb"
	} else if filesize < 1_000_000_000_000 {
		filesize /= 1_000_000_000
		return strconv.Itoa(int(filesize)) + " gb"
	}
	return ""
}
