package main

import (
	"testing"

	"github.com/kmulvey/path"
	"github.com/stretchr/testify/assert"
)

func TestFiles(t *testing.T) {

	var skipDirs, err = makeSkipMap("./skip.txt")
	assert.NoError(t, err)

	skipImages, err := getSkipFiles("../auto/skipcache")
	assert.NoError(t, err)

	upsizedDirs, err := path.List("/home/kmulvey/empyrean/backup/upscayl/", 2, false, path.NewDirEntitiesFilter())
	assert.NoError(t, err)

	findFilesToUpsize(upsizedDirs, "/home/kmulvey/Documents", skipDirs, skipImages)

}
