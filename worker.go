package main

import (
	"bytes"
	"fmt"
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

	for image := range originalImages {
		// image is the abs path
		var upsizedImage = filepath.Base(image)
		upsizedImage = filepath.Join(outputPath, strings.Replace(upsizedImage, filepath.Ext(upsizedImage), ".png", 1))

		log.Trace(cmdPath, " -g ", strconv.Itoa(gpuID), " -n ", " realesrgan-x4plus ", " -i ", image, " -o ", upsizedImage)

		var cmd = exec.Command(cmdPath, "-g", strconv.Itoa(gpuID), "-n", "realesrgan-x4plus", "-i", image, "-o", upsizedImage)
		var out bytes.Buffer
		var stderr bytes.Buffer
		cmd.Stdout = &out
		cmd.Stderr = &stderr

		var start = time.Now()
		var err = cmd.Run()
		if err != nil {
			errors <- fmt.Errorf("error running command, stderr: %s, go err: %w", stderr.String(), err)
		}
		upsizeTime.Set(float64(time.Since(start)))

		upsizedImages <- upsizedImage
	}
}
