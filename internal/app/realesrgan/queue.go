package realesrgan

import (
	"container/list"
	"fmt"
	"reflect"

	"github.com/kmulvey/path"
)

var zeroFile path.Entry

func init() {
	zeroFile, _ = path.NewEntry("./testfiles/zero")
}

// Dedup files based on abs path
// Insert in size order
//   - empty? insert
//   - is it between this and next? insert
func Add(list *list.List, newImage path.Entry) error {

	// init
	if list.Len() == 0 {
		list.PushFront(newImage)
		return nil
	}

	for currElement := list.Front(); currElement != nil; currElement = currElement.Next() {

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
			list.InsertAfter(newImage, currElement)
			break
		} else if newImage.FileInfo.Size() <= currEntry.FileInfo.Size() {
			list.InsertBefore(newImage, currElement)
			break
		} else {
			list.InsertBefore(newImage, currElement)
			break
		}
	}

	return nil
}
