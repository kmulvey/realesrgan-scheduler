package realesrgan

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"unicode"
	"unicode/utf8"

	log "github.com/sirupsen/logrus"
)

/*
THIS CAME FROM WORKER.GO
*/
type ImageConfig struct {
	SourceFile     string
	UpsizedFile    string
	ModelName      string
	RealesrganPath string
	GpuId          uint8
	Remaining      int
	Progress       chan string
}

var upscaylFailedRegex = regexp.MustCompile(`^decode\simage\s.*\sfailed`)

func Upsize(img ImageConfig) error {

	// we need to check if this file has already been upsized
	if _, err := os.Stat(img.UpsizedFile); err == nil {
		return fmt.Errorf("file already exists: %s", img.UpsizedFile)
	}

	if _, err := os.Stat(filepath.Dir(img.UpsizedFile)); errors.Is(err, fs.ErrNotExist) {
		if err := os.MkdirAll(filepath.Dir(img.UpsizedFile), os.ModePerm); err != nil {
			return fmt.Errorf("unable to create upsized directory: %w", err)
		}
	}

	// upsize it !
	var err = runCmdAndCaptureOutput(img.RealesrganPath, img.SourceFile, img.UpsizedFile, img.ModelName, img.GpuId, img.Progress)
	if err != nil {
		return fmt.Errorf("error running upsize command on file %s, err: %w", img.SourceFile, err)
	}

	return nil
}

// runCmdAndCaptureOutput runs the realesrgan command and captures stdout and passes it to logProgress for single line logging.
func runCmdAndCaptureOutput(cmdPath, inputImagePath, upsizedImagePath, modelName string, gpuID uint8, progress chan string) error {

	// these variables were linted up the chain
	//nolint:gosec
	var cmd = exec.Command(cmdPath, "-f", filepath.Ext(upsizedImagePath), "-g", strconv.Itoa(int(gpuID)), "-n", modelName, "-i", inputImagePath, "-o", upsizedImagePath)
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
		errStdout = logProgress(stdoutIn, progress)
		wg.Done()
	}()

	if err := logProgress(stderrIn, progress); err != nil {
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

// logProgress reads the process output and sends the progress (3.5%) on the progress channel and returns any errors.
func logProgress(r io.Reader, progress chan string) error {

	buf := make([]byte, 1024)
	for {
		n, err := r.Read(buf)
		if n > 0 {
			d := buf[:n]
			var rune, _ = utf8.DecodeRune(buf[0:1])
			if unicode.IsDigit(rune) {
				progress <- strings.TrimSpace(string(d))
			} else {
				var line = string(d)
				if strings.Contains(line, "failed") || strings.Contains(line, "Segmentation fault (core dumped)") || upscaylFailedRegex.MatchString(line) {
					return errors.New(line)
				}
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
