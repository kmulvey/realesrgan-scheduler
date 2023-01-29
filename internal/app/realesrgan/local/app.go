package local

import (
	"fmt"

	"github.com/kmulvey/path"
	"github.com/kmulvey/realesrgan-scheduler/internal/cache"
	"github.com/kmulvey/realesrgan-scheduler/internal/fs"
	"github.com/kmulvey/realesrgan-scheduler/internal/queue"
	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
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

	var rl = RealesrganLocal{
		PromNamespace:   promNamespace,
		RealesrganPath:  realesrganPath,
		OutputPath:      outputPath,
		UpsizeTimeGauge: upsizeTime,
		NumGPUs:         numGPUs,
		RemoveOriginals: removeOriginals,
		Queue:           queue.New(false),
	}

	var cache, err = cache.New(cacheDir)
	rl.Cache = cache

	return &rl, err
}

// SetOutputPath allows you to change the output path while RealesrganLocal is running.
func (rl *RealesrganLocal) SetOutputPath(outputPath string) {
	rl.OutputPath = outputPath
}

// Run starts an infinite loop that pulls files from the queue and upsizes them. This can be stopped by calling cancel() on the given context.
func (rl *RealesrganLocal) Run(images []path.Entry) error {

	for _, image := range images {
		var err = rl.AddImage(image)
		if err != nil {
			return fmt.Errorf("problem adding existing files to queue: %w", err)
		}

	}

	rl.UpsizeQueue(0)

	return nil
}

// Watch takes watchEvents and adds them to the queue and listens to events from the queue.
func (rl *RealesrganLocal) Watch(watchEvents chan path.WatchEvent) {

	// start up conversion loop
	var images = make(chan path.Entry)
	rl.UpsizeWatch(rl.NumGPUs, images)

	// listen for events from the queue and when we get one,
	// send NextImage() to the conversion loop
	go func() {
		for range rl.Queue.Notifications {
			images <- rl.Queue.NextImage()
		}
	}()

	// add watch events to the queue. DO NOT add these directly to the conversion loop
	// as that will bypass the ordering of the queue.
	go func() {
		for watchEvent := range watchEvents {
			var err = rl.AddImage(watchEvent.Entry)
			if err != nil {
				log.Errorf("problem adding existing files to queue: %s", err)
			}
		}
	}()

}

// AddImage adds the given image to the queue if the upsized path does not already exist.
func (rl *RealesrganLocal) AddImage(image path.Entry) error {

	if !fs.AlreadyUpsized(image, rl.OutputPath) && !rl.Cache.Contains(image) {
		var err = rl.Queue.Add(image)
		if err != nil {
			return fmt.Errorf("problem adding existing files to queue: %w", err)
		}
	}

	return nil
}
