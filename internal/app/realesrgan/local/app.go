package local

import (
	"context"
	"fmt"

	log "github.com/sirupsen/logrus"

	"github.com/kmulvey/path"
	"github.com/kmulvey/realesrgan-scheduler/internal/queue"
	"github.com/prometheus/client_golang/prometheus"
)

type RealesrganLocal struct {
	PromNamespace   string
	UpsizeTimeGauge prometheus.Gauge
	queue.Queue
}

// NewRealesrganLocal is the constructor for running local upsizing. It takes a slice of existing files
// and prepopulates the queue with them,  Run() takes a channel of watchEvents to stream files.
func NewRealesrganLocal(promNamespace string, existingFiles []path.WatchEvent) (*RealesrganLocal, error) {

	var upsizeTime = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: promNamespace,
			Name:      "upsize_time",
			Help:      "time it tool to upsize the image",
		},
	)
	prometheus.MustRegister(upsizeTime)

	var rl = RealesrganLocal{PromNamespace: promNamespace, UpsizeTimeGauge: upsizeTime}

	rl.Queue = queue.NewQueue()
	for _, existingFile := range existingFiles {
		var err = rl.Queue.Add(existingFile.Entry)
		if err != nil {
			return nil, fmt.Errorf("problem adding existing files to queue: %w", err)
		}
	}

	return &rl, nil
}

// Run starts an infinite loop that pulls files from the queue and upsizes them. This can be stopped by calling cancel() on the given context.
func (rl *RealesrganLocal) Run(ctx context.Context, cmdPath, outputPath string, gpuID int, watchEvents chan path.WatchEvent, errors chan error) {

	go rl.UpsizeWorker(ctx, cmdPath, outputPath, gpuID, errors)

	for {
		select {
		case <-ctx.Done():
			return

		case ev := <-watchEvents:
			var err = rl.Queue.Add(ev.Entry)
			if err != nil {
				errors <- fmt.Errorf("error adding file to queue: %w", err)
			}

		default:
			log.Error(<-errors)
		}
	}
}
