package collection

type Equaler[T any] func(original, new T) bool
type Hasher[T any, C comparable] func(obj T) C

type Pair[K any, V any] struct {
	Key   K
	Value V
}

// Map To avoid Value copy, you may want K, V to be pointer types.
//  However, if Key is a pointer type, we must make sure that the hash code remains the same.
type Map[K any, V any] interface {
	// Collection For the default implementation, Collection.RemoveFirst(pair) equals Remove(pair.Key).
	//  We are not able to determine if the value is equal or not, unless an equaler for value is passed.
	Collection[Pair[K, V]]

	ContainsKey(key K) bool
	Put(key K, value V) (old V, existing bool)
	Get(key K) (value V, existing bool)
	Remove(key K) (old V, existing bool)
}

func NewMap[K any, V any, C comparable](hasher Hasher[K, C], equaler Equaler[K]) Map[K, V] {
	return &mapImpl[K, V, C]{
		data:    map[C][]*Pair[K, V]{},
		hasher:  hasher,
		equaler: equaler,
		size:    0,
	}
}

type mapImpl[K any, V any, C comparable] struct {
	data    map[C][]*Pair[K, V]
	hasher  Hasher[K, C]
	equaler Equaler[K]
	size    int
}

func (m *mapImpl[K, V, C]) Add(pair Pair[K, V]) (oldItem Pair[K, V], replaced bool) {
	oldValue, replaced := m.Put(pair.Key, pair.Value)
	if replaced {
		oldItem.Key = pair.Key
		oldItem.Value = oldValue
		return
	}

	return
}

func (m *mapImpl[K, V, C]) RemoveFirst(pair Pair[K, V]) bool {
	_, exsiting := m.Remove(pair.Key)
	return exsiting
}

func (m *mapImpl[K, V, C]) Has(pair Pair[K, V]) bool {
	return m.ContainsKey(pair.Key)
}

func (m *mapImpl[K, V, C]) Pop() (pair Pair[K, V], existing bool) {
	for _, pairs := range m.data {
		pair = *pairs[len(pairs)-1]
		m.Remove(pair.Key)
		existing = true
		return
	}

	existing = false
	return
}

func (m *mapImpl[K, V, C]) Put(key K, value V) (old V, existing bool) {
	hash := m.hasher(key)
	pairs, exists := m.data[hash]
	if exists {
		for _, pair := range pairs {
			if m.equaler(key, pair.Key) {
				old = pair.Value
				pair.Key = key
				pair.Value = value
				return old, true
			}
		}
		pairs = append(pairs, &Pair[K, V]{
			Key:   key,
			Value: value,
		})
		m.data[hash] = pairs
		m.size += 1
		existing = false
		return
	} else {
		m.data[hash] = []*Pair[K, V]{{
			Key:   key,
			Value: value,
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
		if m.equaler(key, pair.Key) {
			return pair.Value, true
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
		if m.equaler(key, kvPair.Key) {
			if len(pairs) == 1 {
				delete(m.data, hash)
			} else {
				newPairs := pairs[:i]
				newPairs = append(newPairs, pairs[i+1:]...)
				m.data[hash] = newPairs
			}
			m.size -= 1
			return kvPair.Value, true
		}
	}

	existing = false
	return
}

func (m *mapImpl[K, V, C]) Clear() {
	m.data = map[C][]*Pair[K, V]{}
	m.size = 0
}
