package realesrgan

import (
	_ "embed"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestUpsizeFail(t *testing.T) {

	var progress = make(chan string)
	var config = ImageConfig{
		SourceFile:     "/home/kmulvey/Documents/Alexa Bliss/3fus108l7ck61.jpg",
		UpsizedFile:    "/home/kmulvey/empyrean/backup/upscayl/Alexa Bliss/3fus108l7ck61.jpg",
		ModelName:      "realesrgan-x4plus",
		RealesrganPath: "/home/kmulvey/src/realesrgan-ncnn-vulkan-20220424-ubuntu/realesrgan-ncnn-vulkan",
		GpuId:          1,
		Progress:       progress,
	}

	go func() {
		for p := range progress {
			fmt.Println(p)
		}
	}()

	var err = Upsize(config)
	assert.Error(t, err)
	assert.Equal(t, "error running upsize command on file /home/kmulvey/Documents/Alexa Bliss/3fus108l7ck61.jpg, err: error capturing stdErr output: decode image /home/kmulvey/empyrean/backup/upscayl/Alexa Bliss/3fus108l7ck61.jpg failed\n", err.Error())
}

func TestUpsizeSuccess(t *testing.T) {
	t.Parallel()

	var progress = make(chan string)
	var config = ImageConfig{
		SourceFile:     "/home/kmulvey/Documents/Galina Dubenenko/MI6UAzC.jpg",
		UpsizedFile:    "/home/kmulvey/empyrean/backup/upscayl/Galina Dubenenko/MI6UAzC.jpg",
		ModelName:      "realesrgan-x4plus",
		RealesrganPath: "/home/kmulvey/src/realesrgan-ncnn-vulkan-20220424-ubuntu/realesrgan-ncnn-vulkan",
		GpuId:          1,
		Progress:       progress,
	}

	go func() {
		for p := range progress {
			fmt.Println(p)
		}
	}()

	var err = Upsize(config)
	assert.NoError(t, err)
}
