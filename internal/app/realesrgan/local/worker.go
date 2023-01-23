package local

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/kmulvey/path"
	log "github.com/sirupsen/logrus"
)

func (rl RealesrganLocal) UpsizeWorker(cmdPath, outputPath string, gpuID int, originalImages, upsizedImages chan path.Entry, errors chan error) {

	defer close(errors)

	var outputExt = "jpg"

	for image := range originalImages {

		// image is the abs path
		var upsizedImage = image
		upsizedImage.AbsolutePath = filepath.Base(image.AbsolutePath)
		upsizedImage.AbsolutePath = filepath.Join(outputPath, strings.Replace(upsizedImage.AbsolutePath, filepath.Ext(upsizedImage.AbsolutePath), "."+outputExt, 1))

		// we need to check if this file has already been upsized
		if stat, _ := os.Stat(upsizedImage.AbsolutePath); stat != nil {
			var err = os.Remove(image.AbsolutePath)
			if err != nil {
				errors <- fmt.Errorf("error removing original file after upscale, err: %w", err)
			}
			log.WithFields(log.Fields{
				"queue length":  len(originalImages),
				"original":      image.AbsolutePath,
				"original size": prettyPrintFileSizes(image.FileInfo.Size()),
				"upsized":       upsizedImage.AbsolutePath,
			}).Info("already exists, skipping and deleting original")
			continue
		}

		// we log before exec so we can see whats currently being worked on as large files can take several minutes
		log.Trace(cmdPath, "-f", outputExt, " -g ", strconv.Itoa(gpuID), " -n ", " realesrgan-x4plus ", " -i ", image, " -o ", upsizedImage)
		log.WithFields(log.Fields{
			"queue length":  len(originalImages) + 1, // + 1 here because its currently being processed
			"original":      image.AbsolutePath,
			"original size": prettyPrintFileSizes(image.FileInfo.Size()),
		}).Info("upscaling")

		// upsize it !
		var cmd = exec.Command(cmdPath, "-f", outputExt, "-g", strconv.Itoa(gpuID), "-n", "realesrgan-x4plus", "-i", image.AbsolutePath, "-o", upsizedImage.AbsolutePath)
		var out bytes.Buffer
		var stderr bytes.Buffer
		cmd.Stdout = &out
		cmd.Stderr = &stderr

		var start = time.Now()
		var err = cmd.Run()
		if err != nil {
			errors <- fmt.Errorf("error running upsize command on file %s, stderr: %s, stdout: %s, go err: %w", image.AbsolutePath, cleanStdErr(out.String()), cmd.Stdout, err)
			continue
		}
		var duration = time.Since(start)
		rl.UpsizeTimeGauge.Set(float64(duration))

		// if we got here it was successful
		log.WithFields(log.Fields{
			"queue length":  len(originalImages),
			"upsized":       upsizedImage.AbsolutePath,
			"original size": prettyPrintFileSizes(upsizedImage.FileInfo.Size()),
			"duration":      duration,
		}).Info("upsized")

		err = os.Remove(image.AbsolutePath)
		if err != nil {
			errors <- fmt.Errorf("error removing original file after upscale, err: %w", err)
		}

		upsizedImages <- upsizedImage
	}
}

// cleanStdErr removes a lot of progress logging that realesrgan outputs and it not exactly always accurarte, so we discard it for now.
func cleanStdErr(err string) string {

	var builder = strings.Builder{}
	var scanner = bufio.NewScanner(strings.NewReader(err))

	scanner.Split(bufio.ScanLines)

	for scanner.Scan() {
		var line = scanner.Text()
		if !strings.HasSuffix(line, "%") {
			builder.WriteString(line)
		}
	}
	return builder.String()
}

// prettyPrintFileSizes takes file sizes in int and returns a human readable size e.g. "140mb" as a string.
func prettyPrintFileSizes(filesize int64) string {

	if filesize < 1_000 {
		return strconv.Itoa(int(filesize)) + " bytes"
	} else if filesize < 1_000_000 {
		filesize /= 1_000
		return strconv.Itoa(int(filesize)) + " kb"
	} else if filesize < 1_000_000_000 {
		filesize /= 1_000_000
		return strconv.Itoa(int(filesize)) + " mb"
	} else if filesize < 1_000_000_000_000 {
		filesize /= 1_000_000_000
		return strconv.Itoa(int(filesize)) + " gb"
	}
	return ""
}
