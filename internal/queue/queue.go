package queue

import (
	"container/list"
	"fmt"
	"os"
	"reflect"
	"sync"

	"github.com/kmulvey/realesrgan-scheduler/pkg/realesrgan"
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

// New takes a notifications arg which specifies if you want to be notified when a new image is added to the queue.
// If true you must read from Queue.Notifications otherwise it will block Add().
func New(notifications bool) *Queue {

	var q = Queue{List: list.New(), RemovedImages: make(map[string]struct{})}

	if notifications {
		q.Notifications = make(chan struct{})
	}

	return &q
}

// NextImage returns the path.Entry for the image at the front of the queue.
// If there are no more entires it will return an empty path.Entry, as such you will need to check its value.
func (q *Queue) NextImage() *realesrgan.ImageConfig {

	q.Lock.Lock()
	defer q.Lock.Unlock()

	var next = q.List.Front()
	if next == nil {
		return nil
	}

	var nextImage, _ = next.Value.(*realesrgan.ImageConfig) // we dont bother checking if the cast went well because there is no way you could have pushed a non Entry on anyway

	q.RemovedImages[nextImage.SourceFile] = struct{}{}
	q.List.Remove(next)

	return nextImage
}

// Add dedup files based on abs path and adds the given image to the list in size order.
func (q *Queue) Add(newImage *realesrgan.ImageConfig) error {

	q.Lock.Lock()
	defer q.Lock.Unlock()

	// Skip in-flight images
	if _, found := q.RemovedImages[newImage.SourceFile]; found {
		return nil
	}

	// init
	if q.List.Len() == 0 {
		q.List.PushFront(newImage)
		if q.Notifications != nil {
			q.Notifications <- struct{}{}
		}
		return nil
	}

ElementLoop:
	for currElement := q.List.Front(); currElement != nil; currElement = currElement.Next() {

		// this should never happen but we check it for checking's sake
		var currEntry, ok = currElement.Value.(*realesrgan.ImageConfig)
		if !ok {
			return fmt.Errorf("casting currFile to path.Entry failed, was actually type: %s", reflect.TypeOf(currElement.Value))
		}

		// dedup
		if newImage.SourceFile == currEntry.SourceFile {
			return nil

		}

		var hasNext = currElement.Next() != nil
		var newImageFileInfo, err = os.Stat(newImage.SourceFile)
		if err != nil {
			return fmt.Errorf("error getting file info for %s: %w", newImage.SourceFile, err)
		}
		currEntryFileInfo, err := os.Stat(currEntry.SourceFile)
		if err != nil {
			return fmt.Errorf("error getting file info for %s: %w", newImage.SourceFile, err)
		}

		switch {

		case newImageFileInfo.Size() >= currEntryFileInfo.Size() && hasNext:
			continue ElementLoop

		case newImageFileInfo.Size() >= currEntryFileInfo.Size() && !hasNext:
			q.List.InsertAfter(newImage, currElement)
			break ElementLoop

		case newImageFileInfo.Size() <= currEntryFileInfo.Size():
			q.List.InsertBefore(newImage, currElement)
			break ElementLoop

		default:
			q.List.InsertBefore(newImage, currElement)
			break ElementLoop
		}
	}

	if q.Notifications != nil {
		q.Notifications <- struct{}{}
	}

	return nil
}

// Len returns the number of elements of list l. The complexity is O(1).
func (q *Queue) Len() int {
	return q.List.Len()
}

// Contains checks if the targetImage is present in the queue.
// It iterates through the elements of the queue and compares the SourceFile
// field of each element with the SourceFile field of the targetImage.
// If a match is found, it returns true. Otherwise, it returns false.
//
// Parameters:
//
//	targetImage - a pointer to the realesrgan.ImageConfig to be checked.
//
// Returns:
//
//	bool - true if the targetImage is found in the queue, false otherwise.
func (q *Queue) Contains(targetImage *realesrgan.ImageConfig) bool {
	q.Lock.Lock()
	defer q.Lock.Unlock()

	for currElement := q.List.Front(); currElement != nil; currElement = currElement.Next() {

		var currEntry, ok = currElement.Value.(*realesrgan.ImageConfig)
		if !ok {
			log.Errorf("casting currFile to path.Entry failed, was actually type: %s", reflect.TypeOf(currElement.Value))
		}

		if targetImage.SourceFile == currEntry.SourceFile {
			return true
		}
	}

	return false
}

// Print does just that for the whole queue from Front to Back.
func (q *Queue) Print() {

	for currElement := q.List.Front(); currElement != nil; currElement = currElement.Next() {

		var currEntry, ok = currElement.Value.(*realesrgan.ImageConfig)
		if !ok {
			log.Errorf("casting currFile to path.Entry failed, was actually type: %s", reflect.TypeOf(currElement.Value))
		}

		fmt.Println(currEntry.SourceFile)
	}
}
