package semaphore

type Semaphore struct {
	resources chan any
}

func New(resources ...any) *Semaphore {
	var s = &Semaphore{
		resources: make(chan any, len(resources)),
	}

	for _, resource := range resources {
		s.resources <- resource
	}

	return s
}

func (s *Semaphore) Barrow() any {
	return <-s.resources
}

func (s *Semaphore) Return(resource any) {
	s.resources <- resource
}
