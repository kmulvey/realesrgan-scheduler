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
	var searchTerm, addImage string
	var h, ver bool

	flag.Var(&cacheDir, "cache-dir", "where to store the cache file for failed upsizes")
	flag.StringVar(&searchTerm, "search", "", "search term")
	flag.StringVar(&addImage, "add-image", "", "image to add to cache")
	flag.BoolVar(&ver, "version", false, "print version")
	flag.BoolVar(&h, "help", false, "print options")
	flag.Parse()

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

	log.Infof("Config: cacheDir: %s, searchTerm: %s", cacheDir.String(), searchTerm)

	var db, err = cache.New(cacheDir.String())
	if err != nil {
		log.Errorf("error opening badger dir: %s", err)
	}

	addImage = strings.TrimSpace(addImage)
	if addImage != "" {
		var entry, err = path.NewEntry(addImage, 0)
		if err != nil {
			log.Fatalf("image: %s does not exist", addImage)
		}

		err = db.AddImage(entry)
		if err != nil {
			log.Fatalf("error adding image: %s to cache", addImage)
		}

		os.Exit(0)
	}

	var images = make(chan string)
	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		for img := range images {
			fmt.Println(img)
		}
		wg.Done()
	}()

	err = db.ListKeys(searchTerm, images)
	if err != nil {
		log.Errorf("error listing keys: %s", err)
	}

	wg.Wait()
}
