package local

import (
	"github.com/kmulvey/path"
	log "github.com/sirupsen/logrus"
)

// UpsizeQueue upsizes all the images in the queue and returns.
func (rl *RealesrganLocal) UpsizeQueue(gpuID int) {

	var inputImages = make(chan path.Entry)
	rl.UpsizeWatch(inputImages)

	for rl.Queue.Len() > 0 {
		var inputImage = rl.Queue.NextImage()
		inputImages <- inputImage

		log.WithFields(log.Fields{
			"remaining queue length": rl.Queue.Len(),
			"original":               inputImage.AbsolutePath,
			"original size":          PrettyPrintFileSizes(inputImage.FileInfo.Size()),
		}).Info("upscaling")
	}
}

// UpsizeWatch allows for the running of more than one worker thread at once for use with multiple gpus.
func (rl *RealesrganLocal) UpsizeWatch(inputImages chan path.Entry) {
	for i := 0; i <= rl.NumGPUs; i++ {
		go rl.UpsizeLoop(i, inputImages)
	}
}

// UpsizeLoop reads from the images chan and upsizes the image.
func (rl *RealesrganLocal) UpsizeLoop(gpuID int, inputImages chan path.Entry) {
	for image := range inputImages {
		rl.Upsize(image, gpuID)
	}
}
