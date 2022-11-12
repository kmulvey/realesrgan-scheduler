package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/kmulvey/path"
	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
)

func init() {
	prometheus.MustRegister(upsizeTime)
}

var upsizeTime = prometheus.NewGauge(
	prometheus.GaugeOpts{
		Namespace: promNamespace,
		Name:      "upsize_time",
		Help:      "time it tool to upsize the image",
	},
)

func upsizeWorker(cmdPath, outputPath string, gpuID int, originalImages, upsizedImages chan path.Entry, errors chan error) {
	defer close(errors)

	var outputExt = "jpg"

	for image := range originalImages {
		// image is the abs path
		var upsizedImage = image
		upsizedImage.AbsolutePath = filepath.Base(image.AbsolutePath)
		upsizedImage.AbsolutePath = filepath.Join(outputPath, strings.Replace(upsizedImage.AbsolutePath, filepath.Ext(upsizedImage.AbsolutePath), "."+outputExt, 1))

		if stat, _ := os.Stat(upsizedImage.AbsolutePath); stat != nil {
			var err = os.Remove(image.AbsolutePath)
			if err != nil {
				errors <- fmt.Errorf("error removing original file after upscale, err: %w", err)
			}
			log.WithFields(log.Fields{
				"queue length": len(originalImages) + 1,
				"original":     image.AbsolutePath,
				"upsized":      upsizedImage.AbsolutePath,
			}).Info("already exists, skipping and deleting original")
			continue
		}

		log.Trace(cmdPath, "-f", outputExt, " -g ", strconv.Itoa(gpuID), " -n ", " realesrgan-x4plus ", " -i ", image, " -o ", upsizedImage)
		log.WithFields(log.Fields{
			"queue length": len(originalImages) + 1,
			"original":     image.AbsolutePath,
		}).Info("upscaling")

		var cmd = exec.Command(cmdPath, "-f", outputExt, "-g", strconv.Itoa(gpuID), "-n", "realesrgan-x4plus", "-i", image.AbsolutePath, "-o", upsizedImage.AbsolutePath)
		var out bytes.Buffer
		var stderr bytes.Buffer
		cmd.Stdout = &out
		cmd.Stderr = &stderr

		var start = time.Now()
		var err = cmd.Run()
		if err != nil {
			errors <- fmt.Errorf("error running command, stderr: %s, stdout: %s, go err: %w", cmd.Stderr, cmd.Stdout, err)
		}
		upsizeTime.Set(float64(time.Since(start)))

		err = os.Remove(image.AbsolutePath)
		if err != nil {
			errors <- fmt.Errorf("error removing original file after upscale, err: %w", err)
		}

		upsizedImages <- upsizedImage
	}
}
