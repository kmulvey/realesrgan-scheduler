package local

import (
	"sync"

	"github.com/kmulvey/realesrgan-scheduler/pkg/realesrgan"
)

// UpsizeQueue upsizes all the images in the queue using all available gpus.
func (rl *RealesrganLocal) UpsizeQueue() {
	var wg sync.WaitGroup
	var semaphore = make(chan uint8, rl.NumGPUs)
	for i := range rl.NumGPUs {
		semaphore <- i
	}

	for rl.Queue.Len() > 0 {
		var nextImage = rl.Queue.NextImage()
		nextImage.Remaining = rl.Queue.Len() // set the remaining count for the image
		nextImage.GpuId = <-semaphore

		wg.Add(1)
		go func(image *realesrgan.ImageConfig) {
			defer wg.Done()
			defer func() { semaphore <- image.GpuId }() // release the gpu

			rl.files <- image // notify the file is being processed
			realesrgan.Upsize(*image)
		}(nextImage)
	}

	wg.Wait()
}
