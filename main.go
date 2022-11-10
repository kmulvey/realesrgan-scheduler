package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"strconv"

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
	var uploadedImages string
	var upscaledImages string
	var realesrganPath path.Path
	var username string
	var password string
	var port int
	var v bool
	var h bool

	flag.StringVar(&uploadedImages, "uploaded-images-dir", "upload", "where to store the uploaded images")
	flag.StringVar(&upscaledImages, "upscaled-images-dir", "upscaled", "where to store the upscaled images")
	flag.Var(&realesrganPath, "realesrgan-path", "where the realesrgan binary is")
	flag.StringVar(&username, "username", "", "username for the webserver")
	flag.StringVar(&password, "password", "", "password for the webserver")
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

	if err := mkdir(uploadedImages); err != nil {
		log.Fatalf("error creating upload dir: %s", err.Error())
	}

	if err := mkdir(upscaledImages); err != nil {
		log.Fatalf("error creating upscale dir: %s", err.Error())
	}

	var originalImages = make(chan string, 1000)
	var upsizedImages = make(chan string, 1000)

	var app = setupWebServer(originalImages, upsizedImages, uploadedImages, username, password)
	app.Listen(":" + strconv.Itoa(port))
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
