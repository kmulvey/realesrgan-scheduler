package main

import (
	"flag"
	"fmt"
	"path/filepath"
	"regexp"

	"github.com/kmulvey/path"
	"github.com/kmulvey/realesrgan-scheduler/internal/app/realesrgan/local"
)

var jpgRe = regexp.MustCompile(`.*\.jpg$`)

func main() {
	var upscayl, documents string
	flag.StringVar(&upscayl, "original-images-dir", "", "path to the original (input) images")
	flag.StringVar(&documents, "upscaled-images-dir", "", "where to store the upscaled images")
	flag.Parse()

	photoDirs, err := path.List(upscayl, 1, false, path.NewDirEntitiesFilter())
	if err != nil {
		panic(err)
	}

	upscaylDirs := getDirImages(upscayl, photoDirs)
	documentsDirs := getDirImages(documents, photoDirs)

	var diffs = make(map[string]path.Entry)

	for dirName := range upscaylDirs {
		if _, ok := documentsDirs[dirName]; !ok {
			fmt.Println("documents missing dir: ", dirName)
		} else {
			for fileName, fileEntry := range upscaylDirs[dirName] {
				if _, ok := documentsDirs[dirName][fileName]; !ok {
					fmt.Printf("documents/%s missing file: %s\n", dirName, fileName)
					diffs[fileEntry.String()] = fileEntry
				} else {
					delete(documentsDirs[dirName], fileName)
				}
			}

			for _, fileEntry := range documentsDirs[dirName] {
				diffs[fileEntry.String()] = fileEntry
			}
		}
	}

	for filename, entry := range diffs {
		fmt.Printf("Size: %s, Name: %s\n", local.PrettyPrintFileSizes(entry.FileInfo.Size()), filename)
	}
}

func getDirImages(baseDir string, dirEntries []path.Entry) map[string]map[string]path.Entry {
	var dirs = make(map[string]map[string]path.Entry, len(dirEntries))

	for _, dir := range dirEntries {
		var files = make(map[string]path.Entry)
		dirJpgs, err := path.List(filepath.Join(baseDir, dir.FileInfo.Name()), 2, false, path.NewRegexEntitiesFilter(jpgRe))
		if err != nil {
			continue
		}

		for _, entry := range dirJpgs {
			files[entry.FileInfo.Name()] = entry
		}

		dirs[dir.FileInfo.Name()] = files
	}
	return dirs
}
