package local

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPrettyPrintFileSizes(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "100 bytes", PrettyPrintFileSizes(100))
	assert.Equal(t, "1 kb", PrettyPrintFileSizes(1000))
	assert.Equal(t, "10 kb", PrettyPrintFileSizes(10000))
	assert.Equal(t, "100 kb", PrettyPrintFileSizes(100000))
	assert.Equal(t, "1 mb", PrettyPrintFileSizes(1000000))
	assert.Equal(t, "10 mb", PrettyPrintFileSizes(10000000))
	assert.Equal(t, "100 mb", PrettyPrintFileSizes(100000000))
	assert.Equal(t, "987 mb", PrettyPrintFileSizes(987000000))
	assert.Equal(t, "1 gb", PrettyPrintFileSizes(1000000000))
}
