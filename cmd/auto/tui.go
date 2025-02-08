package main

import (
	"github.com/kmulvey/path"
	"github.com/kmulvey/realesrgan-scheduler/internal/app/realesrgan/local"
	"github.com/kmulvey/realesrgan-scheduler/pkg/realesrgan"
	log "github.com/sirupsen/logrus"
)

func tui() {
	var numGPUs = 2

	var skipDirs, err = makeSkipMap("./skip.txt")
	if err != nil {
		log.Fatalf("problem making skip map: %s", err)
	}

	skipImages, err := getSkipFiles("../auto/skipcache")
	if err != nil {
		log.Fatalf("problem getting skip files: %s", err)
	}

	upsizedDirs, err := path.List("/home/kmulvey/empyrean/backup/upscayl/", 2, false, path.NewDirEntitiesFilter())
	if err != nil {
		log.Fatalf("error getting existing upsized dirs: %s", err)
	}

	images, err := findFilesToUpsize(upsizedDirs, "/home/kmulvey/Documents", skipDirs, skipImages)
	if err != nil {
		log.Fatalf("error getting list of new files: %s", err)
	}

	//////////////////
	var files = make(chan *realesrgan.ImageConfig)
	go func() {
		for f := range files {
			log.Infof("processing file: %s", f.SourceFile)
			go func() {
				for pct := range f.Progess {
					log.Infof("%s: %s", f.SourceFile, pct)
				}
			}()
		}
	}()

	rl, err := local.NewRealesrganLocal(promNamespace, "/home/kmulvey/src/realesrgan-ncnn-vulkan-20220424-ubuntu/realesrgan-ncnn-vulkan", "realesrgan-x4plus", uint8(numGPUs), true, files)
	if err != nil {
		log.Fatalf("error in: NewRealesrganLocal %s", err)
	}

	err = rl.Run(images...)
	if err != nil {
		log.Errorf("error in Run(): %s", err)
	}
}
