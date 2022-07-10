package collection

import (
	"container/heap"
)

// Comparator If `first` is less than `second`, then return true
type Comparator[T any] func(first, second T) bool

type PriorityCollection[T any] interface {
	Collection[T]
	Peek() T
}

type PriorityQueue[T any] interface {
	PriorityCollection[T]
}

type PriorityMap[K any, V any] interface {
	PriorityCollection[Pair[K, V]]
	Map[K, V]
}

type PrioritySet[T any] interface {
	PriorityCollection[T]
	Set[T]
}

func NewPriorityQueue[T any](comparator Comparator[T], equaler Equaler[T]) PriorityQueue[T] {
	helper := &priorityHelper[T, emptyType]{
		entries:    []*priorityHelperEntry[T, emptyType]{},
		comparator: comparator,
	}
	heap.Init(helper)
	return &priorityQueue[T]{
		helper:  helper,
		equaler: equaler,
	}
}

func NewPriorityMap[K any, V any, C comparable](
	comparator Comparator[K], hasher Hasher[K, C], equaler Equaler[K]) PriorityMap[K, V] {
	helper := &priorityHelper[K, V]{
		entries:    []*priorityHelperEntry[K, V]{},
		comparator: comparator,
	}
	heap.Init(helper)

	return &priorityMap[K, V]{
		helper:       helper,
		knownEntries: NewMap[K, *priorityHelperEntry[K, V], C](hasher, equaler),
	}
}

func NewPrioritySet[T any, C comparable](
	comparator Comparator[T], hasher Hasher[T, C], equaler Equaler[T]) PrioritySet[T] {
	return &prioritySet[T]{
		set: set[T]{data: NewPriorityMap[T, emptyType, C](comparator, hasher, equaler)},
	}
}

type priorityHelperEntry[K any, V any] struct {
	key   K
	value V
	index int
}

type priorityHelper[K any, V any] struct {
	entries    []*priorityHelperEntry[K, V]
	comparator Comparator[K]
}

func (p *priorityHelper[T, V]) Len() int {
	return len(p.entries)
}

func (p *priorityHelper[T, V]) Less(i, j int) bool {
	return p.comparator(p.entries[i].key, p.entries[j].key)
}

func (p *priorityHelper[T, V]) Swap(i, j int) {
	p.entries[i], p.entries[j] = p.entries[j], p.entries[i]
	p.entries[i].index = i
	p.entries[j].index = j
}

// Push adds an item to the helper. Push should not be called directly; instead,
// use `heap.Push`.
func (p *priorityHelper[T, V]) Push(x any) {
	p.entries = append(p.entries, x.(*priorityHelperEntry[T, V]))
}

// Pop removes an item from the helper. Pop should not be called directly;
// instead, use `heap.Pop`.
func (p *priorityHelper[T, V]) Pop() any {
	n := len(p.entries)
	item := p.entries[n-1]
	item.index = -1
	p.entries = p.entries[0:(n - 1)]
	return item
}

type priorityQueue[T any] struct {
	helper  *priorityHelper[T, emptyType]
	equaler Equaler[T]
}

func (pq *priorityQueue[T]) Has(item T) bool {
	for _, entry := range pq.helper.entries {
		if pq.equaler(item, entry.key) {
			return true
		}
	}
	return false
}

func (pq *priorityQueue[T]) Peek() T {
	if len(pq.helper.entries) == 0 {
		panic("Peek from an empty priorityQueue.")
	}
	return pq.helper.entries[0].key
}

func (pq *priorityQueue[T]) Len() int {
	return pq.helper.Len()
}

func (pq *priorityQueue[T]) Add(item T) (oldItem T, replaced bool) {
	heap.Push(pq.helper, &priorityHelperEntry[T, emptyType]{key: item})
	replaced = false
	return
}

func (pq *priorityQueue[T]) Pop() (item T, existing bool) {
	if pq.Len() <= 0 {
		existing = false
		return
	}

	return heap.Pop(pq.helper).(*priorityHelperEntry[T, emptyType]).key, true
}

func (pq *priorityQueue[T]) RemoveFirst(e T) bool {
	for i, entry := range pq.helper.entries {
		if pq.equaler(e, entry.key) {
			heap.Remove(pq.helper, i)
			return true
		}
	}
	return false
}

func (pq *priorityQueue[T]) Clear() {
	pq.helper.entries = []*priorityHelperEntry[T, emptyType]{}
}

type priorityMap[K any, V any] struct {
	helper       *priorityHelper[K, V]
	knownEntries Map[K, *priorityHelperEntry[K, V]]
}

func (p *priorityMap[K, V]) ContainsKey(key K) bool {
	return p.knownEntries.ContainsKey(key)
}

func (p *priorityMap[K, V]) Put(key K, value V) (old V, existing bool) {
	helperEntry, existing := p.knownEntries.Get(key)

	if existing {
		old = helperEntry.value
		helperEntry.key = key
		helperEntry.value = value
		heap.Fix(p.helper, helperEntry.index)
		// Replace the original one
		p.knownEntries.Put(key, helperEntry)
		return
	} else {
		entry := &priorityHelperEntry[K, V]{key: key, value: value}
		heap.Push(p.helper, entry)
		p.knownEntries.Put(key, entry)
		return
	}
}

func (p *priorityMap[K, V]) Get(key K) (value V, existing bool) {
	helperEntry, existing := p.knownEntries.Get(key)
	if existing {
		value = helperEntry.value
	}
	return
}

func (p *priorityMap[K, V]) Remove(key K) (old V, existing bool) {
	helperEntry, existing := p.knownEntries.Remove(key)
	if existing {
		heap.Remove(p.helper, helperEntry.index)
		old = helperEntry.value
	}

	return
}

func (p *priorityMap[K, V]) Add(pair Pair[K, V]) (oldItem Pair[K, V], replaced bool) {
	oldValue, replaced := p.Put(pair.Key, pair.Value)
	if replaced {
		oldItem.Key = pair.Key
		oldItem.Value = oldValue
		return
	}

	return
}

func (p *priorityMap[K, V]) Pop() (item Pair[K, V], existing bool) {
	if p.Len() <= 0 {
		existing = false
		return
	}

	entry := heap.Pop(p.helper).(*priorityHelperEntry[K, V])
	p.knownEntries.Remove(entry.key)
	item.Key = entry.key
	item.Value = entry.value
	return item, true
}

func (p *priorityMap[K, V]) Peek() (item Pair[K, V]) {
	if len(p.helper.entries) == 0 {
		panic("Peek from an empty priorityQueue.")
	}

	item.Key = p.helper.entries[0].key
	item.Value = p.helper.entries[0].value
	return item
}

func (p *priorityMap[K, V]) Len() int {
	return p.helper.Len()
}

func (p *priorityMap[K, V]) Has(item Pair[K, V]) bool {
	return p.knownEntries.ContainsKey(item.Key)
}

func (p *priorityMap[K, V]) RemoveFirst(item Pair[K, V]) bool {
	_, exsiting := p.Remove(item.Key)
	return exsiting
}

func (pq *priorityMap[K, V]) Clear() {
	pq.helper.entries = []*priorityHelperEntry[K, V]{}
	pq.knownEntries.Clear()
}

type prioritySet[T any] struct {
	set[T]
}

func (s *prioritySet[T]) Peek() T {
	priorityMap := s.set.data.(*priorityMap[T, emptyType])
	return priorityMap.Peek().Key
}
