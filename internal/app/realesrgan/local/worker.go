package local

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/gosuri/uilive"
	"github.com/kmulvey/path"
	"github.com/kmulvey/realesrgan-scheduler/internal/fs"
	log "github.com/sirupsen/logrus"
)

var outputExt = "jpg"

// Upsize does the actual upsizing work.
func (rl *RealesrganLocal) Upsize(inputImage path.Entry, gpuID int) {

	// inputImage is the abs path
	var upsizedImage = inputImage
	upsizedImage.AbsolutePath = filepath.Base(inputImage.AbsolutePath)
	upsizedImage.AbsolutePath = filepath.Join(rl.OutputPath, strings.Replace(upsizedImage.AbsolutePath, filepath.Ext(upsizedImage.AbsolutePath), "."+outputExt, 1))

	// we need to check if this file has already been upsized, this is probably not needed anymore but will require more testing.
	if fs.AlreadyUpsized(inputImage, rl.OutputPath) {
		return
	}

	// we log before exec so we can see whats currently being worked on as large files can take several minutes
	log.Trace(rl.RealesrganPath, "-f", outputExt, " -g ", strconv.Itoa(gpuID), " -n ", " realesrgan-x4plus ", " -i ", inputImage, " -o ", upsizedImage)
	// log.WithFields(log.Fields{
	// 	"remaining queue length": rl.Queue.Len(),
	// 	"original":               inputImage.AbsolutePath,
	// 	"original size":          prettyPrintFileSizes(inputImage.FileInfo.Size()),
	// }).Info("upscaling")

	// upsize it !
	var start = time.Now()
	var err = runCmdAndCaptureOutput(rl.RealesrganPath, outputExt, gpuID, inputImage, upsizedImage)
	if err != nil {
		log.Errorf("error running upsize command on file %s, err: %s", inputImage.AbsolutePath, err)
		err = rl.Cache.AddImage(inputImage)
		if err != nil {
			log.Errorf("error adding broken image to skip cache %s, err: %s", inputImage.AbsolutePath, err)
		}
		return
	}
	var duration = time.Since(start)
	rl.UpsizeTimeGauge.Set(float64(duration))

	// if we got here it was successful
	log.WithFields(log.Fields{
		//"remaining queue length": rl.Queue.Len(),
		"upsized":       upsizedImage.AbsolutePath,
		"original size": prettyPrintFileSizes(upsizedImage.FileInfo.Size()),
		"duration":      duration,
	}).Info("upsized")

	if rl.RemoveOriginals {
		err = os.Remove(inputImage.AbsolutePath)
		if err != nil {
			log.Errorf("error removing original file after upscale, err: %s", err)
		}
	}
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

// runCmdAndCaptureOutput runs the realesrgan command and captures stdout and passes it to logProgress for single line logging.
func runCmdAndCaptureOutput(cmdPath, outputExt string, gpuID int, inputImage, upsizedImage path.Entry) error {

	var cmd = exec.Command(cmdPath, "-f", outputExt, "-g", strconv.Itoa(gpuID), "-n", "realesrgan-x4plus", "-i", inputImage.AbsolutePath, "-o", upsizedImage.AbsolutePath)
	stdoutIn, _ := cmd.StdoutPipe()
	stderrIn, _ := cmd.StderrPipe()
	var errStdout error
	err := cmd.Start()
	if err != nil {
		log.Fatalf("cmd.Start() failed with '%s'\n", err)
	}

	// cmd.Wait() should be called only after we finish reading
	// from stdoutIn and stderrIn.
	// wg ensures that we finish
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		errStdout = logProgress(stdoutIn)
		wg.Done()
	}()

	if err := logProgress(stderrIn); err != nil {
		return fmt.Errorf("error capturing stdErr output: %w", err)
	}

	wg.Wait()
	if errStdout != nil {
		return fmt.Errorf("error capturing stdErr output: %w", err)
	}

	err = cmd.Wait()
	if err != nil {
		return fmt.Errorf("error running cmd: %w", err)
	}
	return nil
}

// logProgress prints the progress loges on a single updating line with uilive.
func logProgress(r io.Reader) error {
	writer := uilive.New()
	// start listening for updates and render
	writer.Start()
	defer writer.Stop()

	buf := make([]byte, 1024)
	for {
		n, err := r.Read(buf[:])
		if n > 0 {
			d := buf[:n]
			var rune, _ = utf8.DecodeRune(buf[0:1])
			if unicode.IsDigit(rune) {
				fmt.Fprintf(writer, "progress: %s \n", strings.TrimSpace(string(d)))
				time.Sleep(time.Millisecond * 10) // required for uilive
			}
		}
		if err != nil {
			// Read returns io.EOF at the end of file, which is not an error for us
			if err == io.EOF {
				err = nil
			}
			return err
		}
	}
}
