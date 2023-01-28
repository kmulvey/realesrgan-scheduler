package queue

import (
	"container/list"
	"fmt"
	"reflect"
	"sync"

	"github.com/kmulvey/path"
	log "github.com/sirupsen/logrus"
)

// Queue is an linked list ordered by file size and deduped.
type Queue struct {
	*list.List
	Lock sync.RWMutex
	// RemovedImages helps us avoid a race condition of adding an image that is currently being processed because it will not be caught by fs.AlreadyExists().
	// A map is used to facilitate thread safety as only having one "CurrImage" would not work with several workers running.
	RemovedImages map[string]struct{}
	// Notifications simply tells you something was added to the queue, you still need to go get it from NextImage(). We dont give you the image here because
	// we dont want to circumvent the Queue's deduping and ordering.
	Notifications chan struct{}
}

// NewQueue takes a notifications arg which specifies if you want to be notified when a new image is added to the queue.
// If true you must read from Queue.Notifications otherwise it will block Add().
func NewQueue(notifications bool) *Queue {

	var q = Queue{List: list.New(), RemovedImages: make(map[string]struct{})}

	if notifications {
		q.Notifications = make(chan struct{})
	}

	return &q
}

// NextImage returns the path.Entry for the image at the front of the queue.
// If there are no more entires it will return an empty path.Entry, as such you will need to check its value.
func (q *Queue) NextImage() path.Entry {

	q.Lock.Lock()
	defer q.Lock.Unlock()

	var next = q.List.Front()
	if next == nil {
		return path.Entry{}
	}

	var nextImage, _ = next.Value.(path.Entry) // we dont bother checking if the cast went well because there is no way you could have pushed a non Entry on anyway

	q.RemovedImages[nextImage.AbsolutePath] = struct{}{}

	return nextImage
}

// Add dedup files based on abs path and adds the given image to the list in size order.
func (q *Queue) Add(newImage path.Entry) error {

	q.Lock.Lock()
	defer q.Lock.Unlock()

	// init
	if q.List.Len() == 0 {
		q.List.PushFront(newImage)
		return nil
	}

	for currElement := q.List.Front(); currElement != nil; currElement = currElement.Next() {

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
			q.List.InsertAfter(newImage, currElement)
			break
		} else if newImage.FileInfo.Size() <= currEntry.FileInfo.Size() {
			q.List.InsertBefore(newImage, currElement)
			break
		} else {
			q.List.InsertBefore(newImage, currElement)
			break
		}
	}

	if q.Notifications != nil {
		q.Notifications <- struct{}{}
	}

	return nil
}

func (q *Queue) Print() {

	for currElement := q.List.Front(); currElement != nil; currElement = currElement.Next() {

		var currEntry, ok = currElement.Value.(path.Entry)
		if !ok {
			log.Errorf("casting currFile to path.Entry failed, was actually type: %s", reflect.TypeOf(currElement.Value))
		}

		fmt.Println(currEntry.AbsolutePath)
	}
}
