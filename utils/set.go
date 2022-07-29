package utils

type Set[T comparable] map[T]struct{}

func (s Set[T]) Add(val T) {
	s[val] = struct{}{}
}

func (s Set[T]) Has(val T) bool {
	_, ok := s[val]

	return ok
}

func (s Set[T]) Delete(val T) {
	delete(s, val)
}

func (s Set[T]) Values() []T {
	vals := make([]T, len(s))

	for val := range s {
		vals = append(vals, val)
	}

	return vals
}
