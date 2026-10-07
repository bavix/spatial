package search

type Output[T any] struct {
	Values  []T
	Visitor func(T) bool
}

func (o *Output[T]) Emit(value T) bool {
	if o.Visitor != nil {
		return o.Visitor(value)
	}

	o.Values = append(o.Values, value)

	return true
}
