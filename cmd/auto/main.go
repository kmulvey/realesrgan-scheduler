package main

import (
	"flag"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/kmulvey/path"
	"github.com/kmulvey/realesrgan-scheduler/internal/app/realesrgan/local"
	"github.com/kmulvey/realesrgan-scheduler/internal/fs"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
	"go.szostok.io/version"
	"go.szostok.io/version/printer"
)

const promNamespace = "realesrgan_scheduler"

func main() {

	log.SetFormatter(&log.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: "2006-01-02 15:04:05",
	})

	go func() {
		http.Handle("/metrics", promhttp.Handler())
		s := &http.Server{
			Addr:           ":6000",
			ReadTimeout:    10 * time.Second,
			WriteTimeout:   10 * time.Second,
			MaxHeaderBytes: 1 << 20,
		}
		log.Fatal(s.ListenAndServe())
	}()

	// get the user options
	var originalImages, upscaledImages, cacheDir path.Path
	var realesrganPath string
	var daemon, removeOriginals, h, ver bool

	flag.Var(&originalImages, "original-images-dir", "path to the original (input) images")
	flag.Var(&upscaledImages, "upscaled-images-dir", "where to store the upscaled images")
	flag.Var(&cacheDir, "cache-dir", "where to store the cache file for failed upsizes")
	flag.StringVar(&realesrganPath, "realesrgan-path", "realesrgan-ncnn-vulkan", "where the realesrgan binary is")
	flag.BoolVar(&removeOriginals, "remove-originals", false, "delete original images after upsizing")
	flag.BoolVar(&daemon, "d", false, "run as a daemon (does not quit)")
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
		originalImages.ComputedPath.AbsolutePath,
		upscaledImages.ComputedPath.AbsolutePath,
		realesrganPath,
		cacheDir.ComputedPath.AbsolutePath,
		removeOriginals,
		daemon)

	var upsizedDirs, err = path.List(upscaledImages.ComputedPath.AbsolutePath, path.NewDirListFilter())
	if err != nil {
		log.Fatalf("error getting existing upsized dirs: %s", err)
	}

	rl, err := local.NewRealesrganLocal(promNamespace, cacheDir.ComputedPath.AbsolutePath, realesrganPath, upscaledImages.ComputedPath.AbsolutePath, 1, removeOriginals)
	if err != nil {
		log.Fatalf("error in: NewRealesrganLocal %s", err)
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	go func() {
		for range c {
			log.Info("interrupt caught, closing db ...")
			err = rl.Cache.Close()
			if err != nil {
				log.Fatalf("error closing badger: %s", err)
			}

			log.Info("bye")
			os.Exit(0)
		}
	}()

	// for each upsized directory, go back to its "originals" dir and look for additional files that have not been upsized.
	for _, upsizedDir := range upsizedDirs {

		var upsizedBase = filepath.Base(upsizedDir.AbsolutePath)
		var originalsDir = filepath.Join(originalImages.ComputedPath.AbsolutePath, upsizedBase)

		rl.SetOutputPath(upsizedDir.AbsolutePath)

		var originalImages, err = path.List(originalsDir, path.NewRegexListFilter(fs.ImageExtensionRegex))

		if err != nil {
			log.Fatalf("error getting existing original images: %s", err)
		}

		log.Infof("Starting queue length: %d for dir: %s", rl.Queue.Len(), originalsDir)

		err = rl.Run(originalImages)
		if err != nil {
			log.Errorf("error in Run(): %s", err)
		}
	}
}
