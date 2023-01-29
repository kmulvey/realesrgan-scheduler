package local

import (
	"github.com/kmulvey/path"
	log "github.com/sirupsen/logrus"
)

// UpsizeQueue upsizes all the images in the queue and returns.
func (rl *RealesrganLocal) UpsizeQueue(gpuID int) {

	for rl.Queue.Len() > 0 {
		var inputImage = rl.Queue.NextImage()

		log.WithFields(log.Fields{
			"remaining queue length": rl.Queue.Len(),
			"original":               inputImage.AbsolutePath,
			"original size":          prettyPrintFileSizes(inputImage.FileInfo.Size()),
		}).Info("upscaling")

		rl.Upsize(inputImage, gpuID)
	}
}

// UpsizeWatch allows for the running of more than one worker thread at once for use with multiple gpus.
func (rl *RealesrganLocal) UpsizeWatch(numGPUs int, inputImages chan path.Entry) {

	for i := 0; i <= numGPUs; i++ {
		// realesrgan has a bug that does not recognize gpu id 1, so it is always skipped
		if i == 1 {
			continue
		}
		go rl.UpsizeLoop(i, inputImages)
	}
}

func (rl *RealesrganLocal) UpsizeLoop(gpuID int, inputImages chan path.Entry) {
	for image := range inputImages {
		rl.Upsize(image, gpuID)
	}
}
