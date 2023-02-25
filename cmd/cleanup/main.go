package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	log "github.com/sirupsen/logrus"
	"go.szostok.io/version"
	"go.szostok.io/version/printer"

	"github.com/kmulvey/path"
	"github.com/kmulvey/realesrgan-scheduler/internal/pkg/ignoreregex"
)

func main() {
	var originalImages path.Path
	var upscaledImages path.Path
	var skipFile string
	var dryRun, v, h bool

	flag.Var(&originalImages, "originals-dir", "")
	flag.Var(&upscaledImages, "upscaled-dir", "")
	flag.StringVar(&skipFile, "skip-file", "", "file with directories to skip, one per line")
	flag.BoolVar(&dryRun, "dry-run", false, "")
	flag.BoolVar(&v, "version", false, "print version")
	flag.BoolVar(&v, "v", false, "print version")
	flag.BoolVar(&h, "help", false, "print options")
	flag.Parse()

	if h {
		flag.PrintDefaults()
		os.Exit(0)
	}

	if v {
		var verPrinter = printer.New()
		var info = version.Get()
		if err := verPrinter.PrintInfo(os.Stdout, info); err != nil {
			log.Fatal(err)
		}
		os.Exit(0)
	}

	for _, dir := range path.FilterEntities(upscaledImages.Files, path.NewDirEntitiesFilter()) {

		var upsizedBase = filepath.Base(dir.AbsolutePath)
		var leafDir = filepath.Join(originalImages.ComputedPath.AbsolutePath, upsizedBase)

		var re, err = ignoreregex.SkipFileToRegexp(skipFile)
		if err != nil {
			log.Fatalf("error creating regex to skip dirs: %s", err)
		}

		if re.MatchString(leafDir) {
			log.Infof("skipping %s ...", dir.FileInfo.Name())
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
		processDir(originalfiles, upsizedfiles, dryRun)
	}
}

func processDir(originalImages, upscaledImages []path.Entry, dryRun bool) {
	var originalsMap = make(map[string]struct{})
	for _, image := range originalImages {
		originalsMap[filepath.Base(image.String())] = struct{}{}
	}
	for _, upscaledImage := range upscaledImages {
		if _, found := originalsMap[filepath.Base(upscaledImage.AbsolutePath)]; !found {
			if !dryRun {
				var err = os.Remove(upscaledImage.AbsolutePath)
				if err != nil {
					log.Fatal(err)
				}
				fmt.Printf("deleted: %s \n", upscaledImage.AbsolutePath)
			} else {
				fmt.Printf("would deleted: %s \n", upscaledImage.AbsolutePath)
			}
		}
	}
}
