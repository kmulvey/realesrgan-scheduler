package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/dgraph-io/badger/v3"
	"github.com/kmulvey/path"
	"github.com/kmulvey/realesrgan-scheduler/internal/fs"
	"github.com/kmulvey/realesrgan-scheduler/pkg/realesrgan"
)

func findFilesToUpsize(upsizeDirs []path.Entry, originalsDir string, skipDirs map[string]struct{}, skipImages map[string]struct{}) ([]realesrgan.ImageConfig, error) {

	var allImages []realesrgan.ImageConfig
	for _, upsizedDir := range upsizeDirs {

		if _, ok := skipDirs[upsizedDir.FileInfo.Name()]; ok {
			continue
		}

		var upsizedBase = filepath.Base(upsizedDir.AbsolutePath)
		var originalsDir = filepath.Join(originalsDir, upsizedBase)

		originalImages, err := path.List(originalsDir, 2, false, path.NewRegexEntitiesFilter(fs.ImageExtensionRegex))
		if err != nil {
			if !errors.Is(err, os.ErrNotExist) {
				return nil, fmt.Errorf("error getting existing original images: %s, err: %w", originalsDir, err)
			}
		}

		var originalsIndex = make(map[string]path.Entry, len(originalImages))
		for _, image := range originalImages {
			if _, ok := skipImages[image.AbsolutePath]; !ok {
				originalsIndex[image.FileInfo.Name()] = image
			}
		}

		existingUpsizedImages, err := path.List(upsizedDir.AbsolutePath, 2, false, path.NewRegexEntitiesFilter(fs.ImageExtensionRegex))
		if err != nil {
			if !errors.Is(err, os.ErrNotExist) {

				return nil, fmt.Errorf("error getting existing upsized images: %s, err: %w", upsizedDir.AbsolutePath, err)
			}
		}

		for _, image := range existingUpsizedImages {
			delete(originalsIndex, image.FileInfo.Name())
		}

		if len(originalsIndex) > 0 {
			for _, entry := range originalsIndex {
				var img = realesrgan.ImageConfig{
					SourceFile:  entry.AbsolutePath,
					UpsizedFile: filepath.Join(upsizedDir.AbsolutePath, entry.FileInfo.Name()),
				}
				allImages = append(allImages, img)
			}
		}
	}

	return allImages, nil
}

func makeSkipMap(skipFile string) (map[string]struct{}, error) {

	var skipMap = make(map[string]struct{})
	readFile, err := os.Open(skipFile)
	if err != nil {
		return nil, fmt.Errorf("error opening skip file: %s, err: %w", skipFile, err)
	}

	fileScanner := bufio.NewScanner(readFile)
	fileScanner.Split(bufio.ScanLines)

	for fileScanner.Scan() {
		skipMap[strings.TrimSpace(fileScanner.Text())] = struct{}{}
	}

	readFile.Close()

	return skipMap, nil
}

func getSkipFiles(dbPath string) (map[string]struct{}, error) {

	db, err := badger.Open(badger.DefaultOptions(dbPath))
	if err != nil {
		return nil, fmt.Errorf("error to open skip file db: %s, err: %w", dbPath, err)
	}
	defer db.Close()

	var images = make(map[string]struct{})
	err = db.View(func(txn *badger.Txn) error {

		var opts = badger.DefaultIteratorOptions
		opts.PrefetchSize = 20
		it := txn.NewIterator(opts)
		defer it.Close()

		for it.Rewind(); it.Valid(); it.Next() {
			images[string(it.Item().Key())] = struct{}{}
		}

		it.Close()
		return nil
	})

	return images, err
}
