package local

import (
	"fmt"

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
	*queue.Queue
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

	rl.Queue = queue.NewQueue(false)

	return &rl, nil
}
func (rl *RealesrganLocal) SetOutputPath(outputPath string) {
	rl.OutputPath = outputPath
}

// Run starts an infinite loop that pulls files from the queue and upsizes them. This can be stopped by calling cancel() on the given context.
func (rl *RealesrganLocal) Run(images []path.Entry) error {

	for _, image := range images {
		if !fs.AlreadyUpsized(image, rl.OutputPath) {
			var err = rl.Queue.Add(image)
			if err != nil {
				return fmt.Errorf("problem adding existing files to queue: %w", err)
			}
		}
	}

	rl.UpsizeQueue(0)

	return nil
}

// AddImage adds the given image to the queue if the upsized path does not already exist.
func (rl *RealesrganLocal) AddImages(images []path.Entry, outputDir string) error {

	for _, image := range images {
		if !fs.AlreadyUpsized(image, outputDir) {
			var err = rl.Queue.Add(image)
			if err != nil {
				return fmt.Errorf("problem adding existing files to queue: %w", err)
			}
		}
	}

	return nil
}

/*
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
*/
