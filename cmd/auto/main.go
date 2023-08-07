package main

import (
	"flag"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/jaypipes/ghw"
	"github.com/kmulvey/path"
	"github.com/kmulvey/realesrgan-scheduler/internal/app/realesrgan/local"
	"github.com/kmulvey/realesrgan-scheduler/internal/fs"
	"github.com/kmulvey/realesrgan-scheduler/internal/pkg/ignoreregex"
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
	var originalImages, upscaledImages, cacheDir path.Entry
	var realesrganPath, skipFile string
	var daemon, removeOriginals, h, ver bool
	var numGPUs int

	flag.Var(&originalImages, "original-images-dir", "path to the original (input) images")
	flag.Var(&upscaledImages, "upscaled-images-dir", "where to store the upscaled images")
	flag.Var(&cacheDir, "cache-dir", "where to store the cache file for failed upsizes")
	flag.StringVar(&realesrganPath, "realesrgan-path", "realesrgan-ncnn-vulkan", "where the realesrgan binary is")
	flag.StringVar(&skipFile, "skip-file", "", "file with directories to skip, one per line")
	flag.IntVar(&numGPUs, "num-gpus", 1, "how many gpus to use")
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

	if gpus, err := getNumGPUs(); err != nil {
		log.Fatal("error getting gpu info: ", err)
	} else if numGPUs > gpus {
		log.Fatalf("cannot use %d gpus as there are only %d.", numGPUs, gpus)
	}

	log.Infof("Config: originalImages: %s, upscaledImages: %s, realesrganPath: %s, cacheDir: %s, skipFIle: %s, removeOriginals: %t, daemon: %t, numGpus: %d",
		originalImages.AbsolutePath,
		upscaledImages.AbsolutePath,
		realesrganPath,
		cacheDir.AbsolutePath,
		skipFile,
		removeOriginals,
		daemon,
		numGPUs)

	var upsizedDirs, err = path.List(upscaledImages.AbsolutePath, 2, false, path.NewDirEntitiesFilter())
	if err != nil {
		log.Fatalf("error getting existing upsized dirs: %s", err)
	}

	rl, err := local.NewRealesrganLocal(promNamespace, cacheDir.AbsolutePath, realesrganPath, upscaledImages.AbsolutePath, numGPUs, removeOriginals, false)
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
		var originalsDir = filepath.Join(originalImages.AbsolutePath, upsizedBase)

		var re, err = ignoreregex.SkipFileToRegexp(skipFile)
		if err != nil {
			log.Fatalf("error creating regex to skip dirs: %s", err)
		}

		if re.MatchString(originalsDir) {
			log.Infof("skipping %s ...", originalsDir)
			continue
		}

		rl.SetOutputPath(upsizedDir.AbsolutePath)

		originalImages, err := path.List(originalsDir, 2, false, path.NewRegexEntitiesFilter(fs.ImageExtensionRegex))
		if err != nil {
			log.Errorf("error getting existing original images: %s", err)
		}

		if len(originalImages) == 0 {
			continue
		}

		log.Infof("Starting queue length: %d for dir: %s", rl.Queue.Len(), originalsDir)

		err = rl.Run(originalImages)
		if err != nil {
			log.Errorf("error in Run(): %s", err)
		}
	}
}

func getNumGPUs() (int, error) {

	var gpu, err = ghw.GPU()
	if err != nil {
		return 0, err
	}

	return len(gpu.GraphicsCards), nil
}
