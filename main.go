package main

import (
	"errors"
	"flag"
	"fmt"
	"os"

	"github.com/kmulvey/path"
	log "github.com/sirupsen/logrus"
	"go.szostok.io/version"
	"go.szostok.io/version/printer"
)

var promNamespace = "realesrgan_scheduler"

func main() {

	log.SetFormatter(&log.TextFormatter{
		FullTimestamp:   true,
		TimestampFormat: "2006-01-02 15:04:05",
	})

	// get the user options
	var inputImages path.Path
	var upscaledImages path.Path
	var realesrganPath string
	var username string
	var password string
	var http bool
	var threads int
	var port int
	var v bool
	var h bool

	flag.Var(&inputImages, "uploaded-images-dir", "where to store the uploaded images")
	flag.Var(&upscaledImages, "upscaled-images-dir", "where to store the upscaled images")
	flag.StringVar(&realesrganPath, "realesrgan-path", "realesrgan-ncnn-vulkan", "where the realesrgan binary is")
	flag.StringVar(&username, "username", "", "username for the webserver")
	flag.StringVar(&password, "password", "", "password for the webserver")
	flag.BoolVar(&http, "http", false, "true to run in webserver mode")
	flag.IntVar(&threads, "threads", 1, "number of gpus")
	flag.IntVar(&port, "port", 3000, "port number for the webserver")
	flag.BoolVar(&v, "version", false, "print version")
	flag.BoolVar(&v, "v", false, "print version")
	flag.BoolVar(&v, "help", false, "print options")
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

	if err := mkdir(inputImages.ComputedPath.AbsolutePath); err != nil {
		log.Fatalf("error creating upload dir: %s", err.Error())
	}

	if err := mkdir(upscaledImages.ComputedPath.AbsolutePath); err != nil {
		log.Fatalf("error creating upscale dir: %s", err.Error())
	}

	var originalImages = make(chan path.Entry, 1000)
	var upsizedImages = make(chan path.Entry, 1000)

	go runWorkers(realesrganPath, upscaledImages.ComputedPath.AbsolutePath, 0, originalImages, upsizedImages)

	if http {
		//var app = setupWebServer(originalImages, upsizedImages, inputImages.ComputedPath.AbsolutePath, username, password)
		//log.WithFields(log.Fields{
		//	"webserver port": port,
		//}).Info("started")
		//app.Listen(":" + strconv.Itoa(port))
	} else {
		go func() {
			for img := range upsizedImages {
				log.WithFields(log.Fields{
					"upsized": img.AbsolutePath,
				}).Info("upsized")
			}
		}()

		log.WithFields(log.Fields{
			"watching directory": inputImages.ComputedPath.AbsolutePath,
		}).Info("started")
		if err := watchDir(inputImages.ComputedPath.AbsolutePath, originalImages); err != nil {
			log.Fatalf("error in watchDir: %s", err)
		}
	}
}

func mkdir(path string) error {
	if _, err := os.Stat(path); errors.Is(err, os.ErrNotExist) {
		err := os.Mkdir(path, os.ModePerm)
		if err != nil {
			return fmt.Errorf("error creating dir: %s, err: %w", path, err)
		}
	}
	return nil
}
