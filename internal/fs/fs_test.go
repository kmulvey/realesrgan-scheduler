package fs

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFindNewImages(t *testing.T) {
	t.Parallel()

	var images, err = FindNewImages("/home/kmulvey/Documents", "/home/kmulvey/empyrean/backup/upscayl", 3)
	assert.NoError(t, err)
	fmt.Println(len(images))

	for _, img := range images {
		fmt.Println(img)
	}

}
