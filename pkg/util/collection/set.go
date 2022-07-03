package collection

// Set To avoid Value copy, you may want T to be pointer types.
//  However, if T is a pointer type, we must make sure that the hash code remains the same.
type Set[T any] interface {
	Collection[T]
}

type emptyType struct{}

var empty emptyType

func NewSet[T any, C comparable](hasher Hasher[T, C], equator Equator[T]) Set[T] {
	return &set[T]{
		data: NewMap[T, emptyType, C](hasher, equator),
	}
}

type set[T any] struct {
	data Map[T, emptyType]
}

func (s *set[T]) Add(item T) {
	s.data.Put(item, empty)
}

func (s *set[T]) RemoveFirst(item T) bool {
	_, exsiting := s.data.Remove(item)
	return exsiting
}

func (s *set[T]) Has(item T) bool {
	return s.data.ContainsKey(item)
}

func (s *set[T]) Pop() (item T, existing bool) {
	pair, existing := s.data.Pop()
	if !existing {
		return
	}

	return pair.Key, existing
}

func (s *set[T]) Len() int {
	return s.data.Len()
}
