package traversal

type Stack[T any] struct {
	items [100]T
	size  int
}

func (s *Stack[T]) Len() int {
	return s.size
}

func (s *Stack[T]) Push(items ...T) {
	copy(s.items[s.size:s.size+len(items)], items)
	s.size += len(items)
}

func (s *Stack[T]) Pop() T {
	s.size--

	return s.items[s.size]
}
