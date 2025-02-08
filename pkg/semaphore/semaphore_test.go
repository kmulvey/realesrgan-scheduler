package semaphore

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSemaphore(t *testing.T) {
	var s = New(1, 2, 3)
	var a = s.Barrow()
	assert.Equal(t, 1, a)
	var b = s.Barrow()
	assert.Equal(t, 2, b)
	var c = s.Barrow()
	assert.Equal(t, 3, c)

	var done = make(chan struct{})
	go func() {
		var d = s.Barrow()
		assert.Equal(t, 1, d)
		close(done)
	}()

	s.Return(1)
	<-done
}
