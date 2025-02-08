package local

import (
	"fmt"

	"github.com/kmulvey/realesrgan-scheduler/internal/queue"
	"github.com/kmulvey/realesrgan-scheduler/pkg/realesrgan"
	"github.com/prometheus/client_golang/prometheus"
)

type RealesrganLocal struct {
	PromNamespace   string
	RealesrganPath  string
	ModelName       string
	NumGPUs         uint8
	RemoveOriginals bool
	UpsizeTimeGauge prometheus.Gauge
	*queue.Queue
	files chan *realesrgan.ImageConfig
}

// NewRealesrganLocal is the constructor for running local upsizing. It takes a slice of existing files
// and prepopulates the queue with them,  Run() takes a channel of watchEvents to stream files.
func NewRealesrganLocal(promNamespace, realesrganPath, modelName string, numGPUs uint8, removeOriginals bool, files chan *realesrgan.ImageConfig) (*RealesrganLocal, error) {

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
		ModelName:       modelName,
		UpsizeTimeGauge: upsizeTime,
		NumGPUs:         numGPUs,
		RemoveOriginals: removeOriginals,
		Queue:           queue.New(false),
		files:           files,
	}

	return &rl, nil
}

// Run starts an infinite loop that pulls files from the queue and upsizes them. This can be stopped by calling cancel() on the given context.
func (rl *RealesrganLocal) Run(images ...*realesrgan.ImageConfig) error {

	for _, image := range images {

		image.ModelName = rl.ModelName
		image.RealesrganPath = rl.RealesrganPath
		image.Progess = make(chan string)

		var err = rl.AddImage(image)
		if err != nil {
			return fmt.Errorf("problem adding existing files to queue: %w", err)
		}
	}

	rl.UpsizeQueue()

	return nil
}

// AddImage adds the given image to the queue if the upsized path does not already exist.
func (rl *RealesrganLocal) AddImage(image *realesrgan.ImageConfig) error {

	if !rl.Queue.Contains(image) {
		var err = rl.Queue.Add(image)
		if err != nil {
			return fmt.Errorf("problem adding existing files to queue: %w", err)
		}
	}

	return nil
}
