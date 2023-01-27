package local

// UpsizeQueue upsizes all the images in the queue and returns.
func (rl *RealesrganLocal) UpsizeQueue(gpuID int) {

	for rl.Queue.Len() > 0 {
		var inputImage = rl.Queue.NextImage()
		rl.Upsize(inputImage, gpuID)
	}
}

// UpsizeWatch allows for the running of more than one worker thread at once for use with multiple gpus.
/* Fan out over multiple GPUs ... table this for now as its not immediately necessary.
func (rl *RealesrganLocal) UpsizeWatch(ctx context.Context, numGPUs int, upsizedImages chan path.Entry) {

	defer close(upsizedImages)
	var errorChans = make([]chan error, numGPUs)

	for i := 0; i <= numGPUs; i++ {
		// realesrgan has a bug that does not recognize gpu id 1, so it is always skipped
		if i == 1 {
			continue
		}
		var errors = make(chan error)
		errorChans[i] = errors
		go rl.UpsizeWorker(ctx, i)
	}

	for err := range goutils.MergeChannels(errorChans...) {
		log.Errorf("error from worker: %s", err.Error())
	}
}
*/
