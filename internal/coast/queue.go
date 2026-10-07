package coast

func (s *Searcher) push(w work) {
	s.queue = append(s.queue, w)
	for child := len(s.queue) - 1; child > 0; {
		parent := (child - 1) / 2
		if s.queue[parent].lower <= s.queue[child].lower {
			return
		}

		s.queue[parent], s.queue[child] = s.queue[child], s.queue[parent]
		child = parent
	}
}

func (s *Searcher) pop() work {
	result := s.queue[0]
	last := len(s.queue) - 1
	s.queue[0] = s.queue[last]
	s.queue[last] = work{}
	s.queue = s.queue[:last]

	for parent := 0; ; {
		child := 2*parent + 1
		if child >= len(s.queue) {
			return result
		}

		if child+1 < len(s.queue) && s.queue[child+1].lower < s.queue[child].lower {
			child++
		}

		if s.queue[parent].lower <= s.queue[child].lower {
			return result
		}

		s.queue[parent], s.queue[child] = s.queue[child], s.queue[parent]
		parent = child
	}
}
