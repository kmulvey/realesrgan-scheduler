package local

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPrettyPrintFileSizes(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "100 bytes", prettyPrintFileSizes(100))
	assert.Equal(t, "1 kb", prettyPrintFileSizes(1000))
	assert.Equal(t, "10 kb", prettyPrintFileSizes(10000))
	assert.Equal(t, "100 kb", prettyPrintFileSizes(100000))
	assert.Equal(t, "1 mb", prettyPrintFileSizes(1000000))
	assert.Equal(t, "10 mb", prettyPrintFileSizes(10000000))
	assert.Equal(t, "100 mb", prettyPrintFileSizes(100000000))
	assert.Equal(t, "987 mb", prettyPrintFileSizes(987000000))
	assert.Equal(t, "1 gb", prettyPrintFileSizes(1000000000))
}
