package queue

/*
func TestAdd(t *testing.T) {
	t.Parallel()

	var queue = New(false)

	var small, err = path.NewEntry("./testfiles/small", 1)
	assert.NoError(t, err)

	medium, err := path.NewEntry("./testfiles/medium", 1)
	assert.NoError(t, err)

	large, err := path.NewEntry("./testfiles/large", 1)
	assert.NoError(t, err)

	assert.NoError(t, queue.Add(small))
	assert.Equal(t, small, queue.List.Front().Value)
	assert.Equal(t, 1, queue.List.Len())

	assert.NoError(t, queue.Add(medium))
	assert.Equal(t, small, queue.List.Front().Value)
	assert.Equal(t, medium, queue.List.Back().Value)
	assert.Equal(t, 2, queue.List.Len())

	assert.NoError(t, queue.Add(large))
	validateAll(t, queue, small, medium, large)

	////////////////////////////////////

	assert.NoError(t, queue.Add(small))
	validateAll(t, queue, small, medium, large)

	assert.NoError(t, queue.Add(large))
	validateAll(t, queue, small, medium, large)

	assert.NoError(t, queue.Add(medium))
	validateAll(t, queue, small, medium, large)

	////////////////////////////////////

	assert.NoError(t, queue.Add(medium))
	validateAll(t, queue, small, medium, large)

	assert.NoError(t, queue.Add(small))
	validateAll(t, queue, small, medium, large)

	assert.NoError(t, queue.Add(large))
	validateAll(t, queue, small, medium, large)

	////////////////////////////////////

	assert.NoError(t, queue.Add(medium))
	validateAll(t, queue, small, medium, large)

	assert.NoError(t, queue.Add(large))
	validateAll(t, queue, small, medium, large)

	assert.NoError(t, queue.Add(small))
	validateAll(t, queue, small, medium, large)

	////////////////////////////////////

	assert.NoError(t, queue.Add(large))
	validateAll(t, queue, small, medium, large)

	assert.NoError(t, queue.Add(small))
	validateAll(t, queue, small, medium, large)

	assert.NoError(t, queue.Add(medium))
	validateAll(t, queue, small, medium, large)

	////////////////////////////////////

	assert.NoError(t, queue.Add(large))
	validateAll(t, queue, small, medium, large)

	assert.NoError(t, queue.Add(medium))
	validateAll(t, queue, small, medium, large)

	assert.NoError(t, queue.Add(small))
	validateAll(t, queue, small, medium, large)

}

// elementAt returns the element from the list at the given position
func elementAt(l *list.List, index int) *list.Element {
	var i int
	for currFile := l.Front(); currFile != nil; currFile = currFile.Next() {
		if i == index {
			return currFile
		}
		i++
	}
	return nil
}

// validateAll runs asserts on the whole list, to reduce code repetition
func validateAll(t *testing.T, queue *Queue, small, medium, large path.Entry) {
	assert.Equal(t, small, queue.List.Front().Value)
	assert.Equal(t, medium, elementAt(queue.List, 1).Value)
	assert.Equal(t, large, queue.List.Back().Value)
	assert.Equal(t, 3, queue.List.Len())
}

// dumper prints the whole list, only use in debugging
// func dumper(l *list.List) {
// 	for currFile := l.Front(); currFile != nil; currFile = currFile.Next() {

// 		var currEntry, _ = currFile.Value.(path.Entry)
// 		fmt.Printf("name: %s, size: %d \n", currEntry.FileInfo.Name(), currEntry.FileInfo.Size())
// 	}
// }
*/
