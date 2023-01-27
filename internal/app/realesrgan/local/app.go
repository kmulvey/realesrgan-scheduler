package local

import (
	"context"
	"fmt"

	log "github.com/sirupsen/logrus"

	"github.com/kmulvey/path"
	"github.com/kmulvey/realesrgan-scheduler/internal/cache"
	"github.com/kmulvey/realesrgan-scheduler/internal/fs"
	"github.com/kmulvey/realesrgan-scheduler/internal/queue"
	"github.com/prometheus/client_golang/prometheus"
)

type RealesrganLocal struct {
	PromNamespace   string
	RealesrganPath  string
	OutputPath      string
	NumGPUs         int
	RemoveOriginals bool
	UpsizeTimeGauge prometheus.Gauge
	queue.Queue
	cache.Cache
}

// NewRealesrganLocal is the constructor for running local upsizing. It takes a slice of existing files
// and prepopulates the queue with them,  Run() takes a channel of watchEvents to stream files.
func NewRealesrganLocal(promNamespace, cacheDir, realesrganPath, outputPath string, numGPUs int, removeOriginals bool) (*RealesrganLocal, error) {

	var upsizeTime = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: promNamespace,
			Name:      "upsize_time",
			Help:      "time it tool to upsize the image",
		},
	)
	prometheus.MustRegister(upsizeTime)

	var rl = RealesrganLocal{PromNamespace: promNamespace, RealesrganPath: realesrganPath, OutputPath: outputPath, UpsizeTimeGauge: upsizeTime, NumGPUs: numGPUs, RemoveOriginals: removeOriginals}

	rl.Queue = queue.NewQueue()

	return &rl, nil
}

// Run starts an infinite loop that pulls files from the queue and upsizes them. This can be stopped by calling cancel() on the given context.
func (rl *RealesrganLocal) Run(ctx context.Context, existingFiles []path.WatchEvent) {

	go rl.UpsizeWorker(ctx, rl.NumGPUs, errors)

	for _, existingFile := range existingFiles {
		var err = rl.Queue.Add(existingFile.Entry)
		if err != nil {
			errors <- fmt.Errorf("problem adding existing files to queue: %w", err)
		}
	}

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

// Run starts an infinite loop that pulls files from the queue and upsizes them. This can be stopped by calling cancel() on the given context.
func (rl *RealesrganLocal) Watch(ctx context.Context, watchEvents chan path.WatchEvent) {

	go rl.UpsizeWorker(ctx, rl.NumGPUs)

	for {
		select {
		case <-ctx.Done():
			return

		case ev := <-watchEvents:
			var err = rl.Queue.Add(ev.Entry)
			if err != nil {
				log.Errorf("error adding file to queue: %s", err)
			}
		}
	}
}

// AddImage adds the given image to the queue if the upsized path does not already exist.
func (rl *RealesrganLocal) AddImage(image path.Entry, outputDir string) error {

	if !fs.AlreadyUpsized(image, outputDir) {
		return rl.Queue.Add(image)
	}

	return nil
}
