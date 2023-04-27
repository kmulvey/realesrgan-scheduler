package local

import (
	"sync"

	"github.com/kmulvey/path"
	log "github.com/sirupsen/logrus"
)

// UpsizeQueue upsizes all the images in the queue and returns.
func (rl *RealesrganLocal) UpsizeQueue() {

	var wg sync.WaitGroup
	var inputImages = make(chan path.Entry)
	rl.UpsizeWatch(&wg, inputImages)

	for rl.Queue.Len() > 0 {
		var inputImage = rl.Queue.NextImage()
		inputImages <- inputImage
		wg.Add(1)

		log.WithFields(log.Fields{
			"remaining queue length": rl.Queue.Len(),
			"original":               inputImage.AbsolutePath,
			"original size":          PrettyPrintFileSizes(inputImage.FileInfo.Size()),
		}).Info("upscaling")
	}

	wg.Wait()
}

// UpsizeWatch allows for the running of more than one worker thread at once for use with multiple gpus.
func (rl *RealesrganLocal) UpsizeWatch(wg *sync.WaitGroup, inputImages chan path.Entry) {
	for i := 0; i <= rl.NumGPUs; i++ {
		i := i
		go rl.UpsizeLoop(wg, i, inputImages)
	}
}

// UpsizeLoop reads from the images chan and upsizes the image.
func (rl *RealesrganLocal) UpsizeLoop(wg *sync.WaitGroup, gpuID int, inputImages chan path.Entry) {
	for image := range inputImages {
		rl.Upsize(wg, image, gpuID)
	}
}
