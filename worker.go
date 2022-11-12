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

func upsizeWorker(cmdPath, outputPath string, gpuID int, originalImages, upsizedImages chan string, errors chan error) {
	defer close(errors)

	var outputExt = "jpg"

	for image := range originalImages {
		// image is the abs path
		var upsizedImage = filepath.Base(image)
		upsizedImage = filepath.Join(outputPath, strings.Replace(upsizedImage, filepath.Ext(upsizedImage), "."+outputExt, 1))

		if stat, _ := os.Stat(upsizedImage); stat != nil {
			var err = os.Remove(image)
			if err != nil {
				errors <- fmt.Errorf("error removing original file after upscale, err: %w", err)
			}
			log.WithFields(log.Fields{
				"queue length": len(originalImages),
				"original":     image,
				"upsized":      upsizedImage,
			}).Info("already exists, skipping and deleting original")
			continue
		}

		log.Trace(cmdPath, "-f", outputExt, " -g ", strconv.Itoa(gpuID), " -n ", " realesrgan-x4plus ", " -i ", image, " -o ", upsizedImage)
		log.WithFields(log.Fields{
			"queue length": len(originalImages),
			"original":     image,
		}).Info("upscaling")

		var cmd = exec.Command(cmdPath, "-f", outputExt, "-g", strconv.Itoa(gpuID), "-n", "realesrgan-x4plus", "-i", image, "-o", upsizedImage)
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

		err = os.Remove(image)
		if err != nil {
			errors <- fmt.Errorf("error removing original file after upscale, err: %w", err)
		}

		upsizedImages <- upsizedImage
	}
}
