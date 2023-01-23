package local

import (
	"github.com/kmulvey/goutils"
	"github.com/kmulvey/path"
	log "github.com/sirupsen/logrus"
)

func (rl RealesrganLocal) RunWorkers(cmdPath, outputPath string, numGPUs int, originalImages, upsizedImages chan path.Entry) {
	defer close(upsizedImages)
	var errorChans = make([]chan error, numGPUs+1)
	for i := 0; i <= numGPUs; i++ {
		// realesrgan has a bug that does not recognize gpu id 1, so it is always skipped
		var errors = make(chan error)
		errorChans[i] = errors
		go rl.UpsizeWorker(cmdPath, outputPath, i, originalImages, upsizedImages, errors)
	}

	for err := range goutils.MergeChannels(errorChans...) {
		log.Errorf("error from worker: %s", err.Error())
	}
}
