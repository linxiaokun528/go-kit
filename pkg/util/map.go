package util

type Equator[T any] func(original, new T) bool
type Hasher[T any, C comparable] func(obj T) C

type pair[K any, V any] struct {
	key   K
	value V
}

// Map To avoid value copy, you may want K, V to be pointer types.
//  However, if key is a pointer type, we must make sure that the hash code remains the same.
type Map[K any, V any] interface {
	Len() int
	ContainsKey(key K) bool
	Put(key K, value V) (old V, existing bool)
	Get(key K) (value V, existing bool)
	Remove(key K) (old V, existing bool)
}

func NewMap[K any, V any, C comparable](hasher Hasher[K, C], equator Equator[K]) Map[K, V] {
	return &mapImpl[K, V, C]{
		data:    map[C][]*pair[K, V]{},
		hasher:  hasher,
		equator: equator,
		size:    0,
	}
}

type mapImpl[K any, V any, C comparable] struct {
	data    map[C][]*pair[K, V]
	hasher  Hasher[K, C]
	equator Equator[K]
	size    int
}

func (m *mapImpl[K, V, C]) Put(key K, value V) (old V, existing bool) {
	hash := m.hasher(key)
	pairs, exists := m.data[hash]
	if exists {
		for _, pair := range pairs {
			if m.equator(key, pair.key) {
				old = pair.value
				pair.key = key
				pair.value = value
				return old, true
			}
		}
		pairs = append(pairs, &pair[K, V]{
			key:   key,
			value: value,
		})
		m.data[hash] = pairs
		m.size += 1
		existing = false
		return
	} else {
		m.data[hash] = []*pair[K, V]{{
			key:   key,
			value: value,
		}}
		m.size += 1
		existing = false
		return
	}
}

func (m *mapImpl[K, V, C]) Get(key K) (value V, existing bool) {
	hash := m.hasher(key)
	pairs, existing := m.data[hash]
	if !existing {
		return
	}

	for _, pair := range pairs {
		if m.equator(key, pair.key) {
			return pair.value, true
		}
	}
	existing = false
	return
}

func (m *mapImpl[K, V, C]) Len() int {
	return m.size
}

func (m *mapImpl[K, V, C]) ContainsKey(key K) bool {
	_, existing := m.Get(key)
	return existing
}

func (m *mapImpl[K, V, C]) Remove(key K) (old V, existing bool) {
	hash := m.hasher(key)
	pairs, existing := m.data[hash]

	if !existing {
		existing = false
		return
	}

	for i, kvPair := range pairs {
		if m.equator(key, kvPair.key) {
			if len(pairs) == 1 {
				delete(m.data, hash)
			} else {
				newPairs := pairs[:i]
				newPairs = append(newPairs, pairs[i+1:]...)
				m.data[hash] = newPairs
			}
			m.size -= 1
			return kvPair.value, true
		}
	}

	existing = false
	return
}
