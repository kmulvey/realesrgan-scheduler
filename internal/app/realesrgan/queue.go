package realesrgan

import (
	"container/list"
	"fmt"
	"reflect"
	"sync"

	"github.com/kmulvey/path"
)

type Queue struct {
	Queue     *list.List
	QueueLock sync.RWMutex
}

func NewQueue() Queue {
	return Queue{Queue: list.New()}
}

// NextImage returns the path.Entry for the image at the front of the queue.
// If there are no more entires it will return an empty path.Entry, as such you will need to check its value.
func (q *Queue) NextImage() path.Entry {

	q.QueueLock.Lock()
	defer q.QueueLock.Unlock()

	var next = q.Queue.Front()
	if next == nil {
		return path.Entry{}
	}

	var nextImage, _ = next.Value.(path.Entry) // we dont bother checking if the cast went well because there is no way you could have pushed a non Entry on anyway

	return nextImage
}

// Add dedup files based on abs path and adds the given image to the list in size order.
func (q *Queue) Add(newImage path.Entry) error {

	q.QueueLock.Lock()
	defer q.QueueLock.Unlock()

	// init
	if q.Queue.Len() == 0 {
		q.Queue.PushFront(newImage)
		return nil
	}

	for currElement := q.Queue.Front(); currElement != nil; currElement = currElement.Next() {

		// this should never happen but we check it for checking's sake
		var currEntry, ok = currElement.Value.(path.Entry)
		if !ok {
			return fmt.Errorf("casting currFile to path.Entry failed, was actually type: %s", reflect.TypeOf(currElement.Value))
		}

		// dedup
		if newImage.AbsolutePath == currEntry.AbsolutePath {
			return nil

		}

		var hasNext = currElement.Next() != nil

		if newImage.FileInfo.Size() >= currEntry.FileInfo.Size() && hasNext {
			continue
		} else if newImage.FileInfo.Size() >= currEntry.FileInfo.Size() && !hasNext {
			q.Queue.InsertAfter(newImage, currElement)
			break
		} else if newImage.FileInfo.Size() <= currEntry.FileInfo.Size() {
			q.Queue.InsertBefore(newImage, currElement)
			break
		} else {
			q.Queue.InsertBefore(newImage, currElement)
			break
		}
	}

	return nil
}
