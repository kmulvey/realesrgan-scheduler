package main

import (
	"testing"

	"github.com/kmulvey/path"
	"github.com/kmulvey/realesrgan-scheduler/internal/app/realesrgan/local"
	"github.com/kmulvey/realesrgan-scheduler/pkg/realesrgan"
	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestFiles(t *testing.T) {

	var skipDirs, err = makeSkipMap("./skip.txt")
	assert.NoError(t, err)

	skipImages, err := getSkipFiles("../auto/skipcache")
	assert.NoError(t, err)

	upsizedDirs, err := path.List("/home/kmulvey/empyrean/backup/upscayl", 2, false, path.NewDirEntitiesFilter())
	assert.NoError(t, err)

	images, err := findFilesToUpsize(upsizedDirs, "/home/kmulvey/Documents", skipDirs, skipImages)
	assert.NoError(t, err)

	//////////////////
	var files = make(chan *realesrgan.ImageConfig)
	go func() {
		for f := range files {
			log.Infof("processing file: %s, remaining: %d", f.SourceFile, f.Remaining)
			go func() {
				for pct := range f.Progress {
					log.Infof("%s: %s", f.SourceFile, pct)
				}
			}()
		}
	}()

	rl, err := local.NewRealesrganLocal(promNamespace, "/home/kmulvey/src/realesrgan-ncnn-vulkan-20220424-ubuntu/realesrgan-ncnn-vulkan", "realesrgan-x4plus", 2, true, files)
	assert.NoError(t, err)

	err = rl.Run(images...)
	assert.NoError(t, err)

}
