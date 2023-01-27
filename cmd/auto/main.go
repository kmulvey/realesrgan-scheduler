package auto

import (
	"context"
	"errors"
	"flag"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/kmulvey/path"
	"github.com/kmulvey/realesrgan-scheduler/internal/app/realesrgan/local"
	"github.com/kmulvey/realesrgan-scheduler/internal/fs"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
	"go.szostok.io/version"
	"go.szostok.io/version/printer"
)

const promNamespace = "realesrgan-scheduler"

func main() {

	var ctx, cancel = context.WithCancel(context.Background())

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
	var inputImages path.Path
	var upscaledImages path.Path
	var realesrganPath string
	var daemon bool
	var v bool
	var h bool

	flag.Var(&inputImages, "uploaded-images-dir", "where to store the uploaded images")
	flag.Var(&upscaledImages, "upscaled-images-dir", "where to store the upscaled images")
	flag.StringVar(&realesrganPath, "realesrgan-path", "realesrgan-ncnn-vulkan", "where the realesrgan binary is")
	flag.BoolVar(&daemon, "d", false, "run as a daemon (does not quit)")
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

	if _, err := os.Stat(inputImages.ComputedPath.AbsolutePath); errors.Is(err, os.ErrNotExist) {
		log.Fatalf("input dir does not exist: %s", err.Error())
	}

	if err := fs.MakeDir(upscaledImages.ComputedPath.AbsolutePath); err != nil {
		log.Fatalf("error creating upscale dir: %s", err.Error())
	}

	var upsizedDirs, err = path.List(inputImages.ComputedPath.AbsolutePath, path.NewDirListFilter())
	if err != nil {
		log.Fatalf("error getting existing upsized dirs: %s", err)
	}

	for _, upsizedDir := range upsizedDirs {

		var upsizedBase = filepath.Base(upsizedDir.AbsolutePath)
		var originalsDir = filepath.Join(inputImages.ComputedPath.AbsolutePath, upsizedBase)

		var originalImages, err = fs.GetExistingFiles(originalsDir)
		if err != nil {
			log.Fatalf("error getting existing original images: %s", err)
		}

		rl, err := local.NewRealesrganLocal(promNamespace, originalImages)
		if err != nil {
			log.Fatalf("error in: NewRealesrganLocal %s", err)
		}

		var errors = make(chan error)
		rl.Run(ctx, realesrganPath, upscaledImages.ComputedPath.AbsolutePath, 0, nil, errors)
	}

	cancel()
}
