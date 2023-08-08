package main

import (
	"context"
	"flag"
	"net/http"
	"os"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/kmulvey/path"
	"github.com/kmulvey/realesrgan-scheduler/internal/app/realesrgan/local"
	"github.com/kmulvey/realesrgan-scheduler/internal/fs"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
	"go.szostok.io/version"
	"go.szostok.io/version/printer"
)

var promNamespace = "realesrgan_scheduler"

func main() {

	var ctx, cancel = context.WithCancel(context.Background())

	log.SetFormatter(&log.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: "2006-01-02 15:04:05",
	})

	go func() {
		http.Handle("/metrics", promhttp.Handler())
		s := &http.Server{
			Addr:           ":6060",
			ReadTimeout:    10 * time.Second,
			WriteTimeout:   10 * time.Second,
			MaxHeaderBytes: 1 << 20,
		}
		log.Fatal(s.ListenAndServe())
	}()

	// get the user options
	var originalImages, upscaledImages, cacheDir path.Entry
	var realesrganPath string
	var daemon, removeOriginals, h, ver bool
	var numGPUs int

	flag.Var(&originalImages, "original-images-dir", "path to the original (input) images")
	flag.Var(&upscaledImages, "upscaled-images-dir", "where to store the upscaled images")
	flag.Var(&cacheDir, "cache-dir", "where to store the cache file for failed upsizes")
	flag.StringVar(&realesrganPath, "realesrgan-path", "realesrgan-ncnn-vulkan", "where the realesrgan binary is")
	flag.BoolVar(&removeOriginals, "remove-originals", false, "delete original images after upsizing")
	flag.BoolVar(&daemon, "d", false, "run as a daemon (does not quit)")
	flag.IntVar(&numGPUs, "num-gpus", 1, "how many gpus to use")
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

	log.Infof("Config: originalImages: %s, upscaledImages: %s, realesrganPath: %s, cacheDir: %s, removeOriginals: %t, daemon: %t",
		originalImages.String(),
		upscaledImages.String(),
		realesrganPath,
		cacheDir.String(),
		removeOriginals,
		daemon)

	var images, err = path.List(originalImages.String(), 1, false, path.NewRegexEntitiesFilter(fs.ImageExtensionRegex))
	if err != nil {
		log.Fatalf("error getting existing upsized dirs: %s", err)
	}

	rl, err := local.NewRealesrganLocal(promNamespace, cacheDir.String(), realesrganPath, upscaledImages.String(), numGPUs, removeOriginals, daemon)
	if err != nil {
		log.Fatalf("error in: NewRealesrganLocal %s", err)
	}

	if daemon {
		var watchEvents = make(chan path.WatchEvent)
		go rl.Watch(watchEvents)

		// load up existing images
		for _, image := range images {
			err = rl.AddImage(image)
			if err != nil {
				log.Fatalf("error adding image to queue: %s", err)
			}
		}

		var errors = make(chan error)
		go func() {
			for err := range errors {
				log.Errorf("error from WatchDir, %s", err)
			}
		}()

		path.WatchDir(ctx, originalImages.String(), 2, false, watchEvents, errors, path.NewOpWatchFilter(fsnotify.Create), path.NewRegexWatchFilter(fs.ImageExtensionRegex))

	} else {
		// load up existing images
		for _, image := range images {
			err = rl.AddImage(image)
			if err != nil {
				log.Fatalf("error adding image to queue: %s", err)
			}
		}

		err = rl.Run(nil) // images were already added above
		if err != nil {
			log.Errorf("error in Run(): %s", err)
		}
	}

	cancel()
}
