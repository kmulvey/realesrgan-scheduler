package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"sync"

	log "github.com/sirupsen/logrus"

	"github.com/kmulvey/path"
	"github.com/kmulvey/realesrgan-scheduler/internal/cache"
	"go.szostok.io/version"
	"go.szostok.io/version/printer"
)

func main() {
	// get the user options
	var cacheDir path.Entry
	var searchTerm, addImage, removeImage string
	var listKeys, purge, h, ver bool

	flag.Var(&cacheDir, "cache-dir", "where to store the cache file for failed upsizes")
	flag.StringVar(&searchTerm, "search", "", "search term")
	flag.StringVar(&addImage, "add-image", "", "image to add to cache")
	flag.StringVar(&removeImage, "remove-image", "", "remove image from cache")
	flag.BoolVar(&listKeys, "list-keys", false, "list all keys")
	flag.BoolVar(&purge, "purge", false, "delete all keys")
	flag.BoolVar(&ver, "version", false, "print version")
	flag.BoolVar(&h, "help", false, "print options")
	flag.Parse()

	addImage = strings.TrimSpace(addImage)
	removeImage = strings.TrimSpace(removeImage)
	searchTerm = strings.TrimSpace(searchTerm)

	if h {
		flag.PrintDefaults()
		os.Exit(0)
	}

	if ver {
		var verPrinter = printer.New()
		var info = version.Get()
		if err := verPrinter.PrintInfo(os.Stdout, info); err != nil {
			log.Fatal(err)
		}
		os.Exit(0)
	}

	var db, err = cache.New(cacheDir.String())
	if err != nil {
		log.Errorf("error opening badger dir: %s", err)
	}

	if addImage != "" {
		if err := add(addImage, db); err != nil {
			log.Errorf("error adding key: %s", err)
			os.Exit(1)
		}
		os.Exit(0)

	} else if removeImage != "" {
		if err = db.RemoveImage(removeImage); err != nil {
			log.Errorf("error removing image: %s from cache, err :%s", removeImage, err)
			os.Exit(1)
		}
		log.Infof("removed :%s from cache", removeImage)
		os.Exit(0)

	} else if purge {
		if err = db.RemoveAllImages(); err != nil {
			log.Errorf("error purging database, err :%s", err)
			os.Exit(1)
		}
		log.Info("purged database")
		os.Exit(0)

	} else if searchTerm != "" || listKeys {
		var results, err = searchImages(searchTerm, db)
		if err != nil {
			log.Errorf("error searching keys: %s", err)
			os.Exit(1)
		}

		for _, img := range results {
			fmt.Println(img)
		}
		os.Exit(0)
	}
}

func add(image string, db cache.Cache) error {
	var entry, err = path.NewEntry(image, 0)
	if err != nil {
		return fmt.Errorf("image: %s does not exist, err :%w", image, err)
	}

	err = db.AddImage(entry)
	if err != nil {
		return fmt.Errorf("error adding image: %s to cache, err :%w", image, err)
	}

	return nil
}

func searchImages(searchTerm string, db cache.Cache) ([]string, error) {
	var images = make(chan string)
	var searchResults = make([]string, 0)
	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		for img := range images {
			searchResults = append(searchResults, img)
		}
		wg.Done()
	}()

	if err := db.ListKeys(searchTerm, images); err != nil {
		log.Errorf("error listing keys: %s", err)
	}

	wg.Wait()

	return searchResults, nil
}
