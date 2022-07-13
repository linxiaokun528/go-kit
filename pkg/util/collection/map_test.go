package collection_test

import (
	"fmt"
	"reflect"
	"strconv"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "go-kit/pkg/util/collection"
)

func basicHasher[K comparable](value K) K {
	return value
}

func basicEquator[K comparable](first, second K) bool {
	return first == second
}

func fakeComparator[K comparable](_, _ K) bool {
	return true
}

type fromInt[T comparable] func(int) T

func testBasicTypesForMap[T comparable](mapType mapType, convert fromInt[T]) {
	var obj T
	typeName := reflect.TypeOf(obj).Name()

	Describe(fmt.Sprintf("It can work with type [%s].", typeName), func() {
		var mapForTest Map[T, T]

		BeforeEach(func() {
			// use a fake comparator here. We don't really need a comparator in this test
			mapForTest = createMap[T, T, T](mapType, basicHasher[T], basicEquator[T], fakeComparator[T])
		})

		It("returns 'existing=false' when the Key is nonexistent.", func() {
			_, existing := mapForTest.Get(convert(0))
			Expect(existing).To(BeFalse())

			mapForTest.Put(convert(1), convert(2))
			_, existing = mapForTest.Get(convert(0))
			Expect(existing).To(BeFalse())
		})

		It("can get what it puts.", func() {
			mapForTest.Put(convert(1), convert(2))
			value, existing := mapForTest.Get(convert(1))
			Expect(value).To(Equal(convert(2)))
			Expect(existing).To(BeTrue())

			// If there are multiple keys, won't get a wrong Value
			mapForTest.Put(convert(0), convert(0))
			value, existing = mapForTest.Get(convert(1))
			Expect(value).To(Equal(convert(2)))
			Expect(existing).To(BeTrue())
		})

		It("can overwrite what it puts.", func() {
			mapForTest.Put(convert(1), convert(0))
			mapForTest.Put(convert(1), convert(3))
			value, existing := mapForTest.Get(convert(1))
			Expect(value).To(Equal(convert(3)))
			Expect(existing).To(BeTrue())
		})

		It("can indicate if the Key is already existent when it puts. If so, it returns the old Value.", func() {
			mapForTest.Put(convert(1), convert(0))
			oldValue, existing := mapForTest.Put(convert(1), convert(1))
			Expect(oldValue).To(Equal(convert(0)))
			Expect(existing).To(BeTrue())

			_, existing = mapForTest.Put(convert(0), convert(3))
			Expect(existing).To(BeFalse())
		})

		It("can remove what it puts.", func() {
			mapForTest.Put(convert(0), convert(2))
			_, existing := mapForTest.Remove(convert(2))
			Expect(existing).To(BeFalse())
			value, existing := mapForTest.Remove(convert(0))
			Expect(existing).To(BeTrue())
			Expect(value).To(Equal(convert(2)))

			mapForTest.Put(convert(1), convert(0))
			mapForTest.Put(convert(1), convert(3))
			value, existing = mapForTest.Remove(convert(1))
			Expect(existing).To(BeTrue())
			Expect(value).To(Equal(convert(3)))
		})

		It("can clear what it puts.", func() {
			mapForTest.Put(convert(0), convert(2))
			mapForTest.Put(convert(1), convert(0))
			mapForTest.Clear()
			Expect(mapForTest.Len()).To(Equal(0))

			mapForTest.Put(convert(0), convert(2))
			Expect(mapForTest.ContainsKey(convert(0))).To(BeTrue())
			Expect(mapForTest.ContainsKey(convert(1))).To(BeFalse())
		})

		It("can return what it puts into an array.", func() {
			Expect(mapForTest.ToArray()).To(BeEmpty())

			pair1 := Pair[T, T]{Key: convert(0), Value: convert(2)}
			pair2 := Pair[T, T]{Key: convert(1), Value: convert(0)}
			mapForTest.Add(pair1)
			mapForTest.Add(pair2)
			Expect(mapForTest.ToArray()).To(ConsistOf(pair1, pair2))

			mapForTest.RemoveFirst(pair1)
			Expect(mapForTest.ToArray()).To(ConsistOf(pair2))
			mapForTest.RemoveFirst(pair2)
			Expect(mapForTest.ToArray()).To(BeEmpty())
			mapForTest.RemoveFirst(pair2)
			Expect(mapForTest.ToArray()).To(BeEmpty())
		})

		It("can show if it contains a specified Key.", func() {
			Expect(mapForTest.ContainsKey(convert(1))).To(BeFalse())
			mapForTest.Put(convert(1), convert(2))
			Expect(mapForTest.ContainsKey(convert(1))).To(BeTrue())
			mapForTest.Put(convert(1), convert(0))
			Expect(mapForTest.ContainsKey(convert(1))).To(BeTrue())
			mapForTest.Remove(convert(0))
			Expect(mapForTest.ContainsKey(convert(1))).To(BeTrue())
			mapForTest.Remove(convert(1))
			Expect(mapForTest.ContainsKey(convert(1))).To(BeFalse())

			// Make sure the "Value" won't make containsKey returns true
			mapForTest.Put(convert(0), convert(3))
			Expect(mapForTest.ContainsKey(convert(3))).To(BeFalse())
			// Make sure it works when it contains multiple keys
			mapForTest.Put(convert(1), convert(0))
			Expect(mapForTest.ContainsKey(convert(1))).To(BeTrue())
		})

		It("can return the number of items it contains.", func() {
			Expect(mapForTest.Len()).To(Equal(0))
			mapForTest.Put(convert(1), convert(2))
			Expect(mapForTest.Len()).To(Equal(1))
			mapForTest.Put(convert(1), convert(0))
			Expect(mapForTest.Len()).To(Equal(1))
			mapForTest.Put(convert(0), convert(3))
			Expect(mapForTest.Len()).To(Equal(2))
			mapForTest.Remove(convert(1))
			Expect(mapForTest.Len()).To(Equal(1))
			mapForTest.Remove(convert(3))
			Expect(mapForTest.Len()).To(Equal(1))
			mapForTest.Remove(convert(0))
			Expect(mapForTest.Len()).To(Equal(0))
		})
	})
}

type idValue struct {
	id    int
	value int
}

func (i *idValue) hash() int {
	return i.id
}

func (i *idValue) equals(other *idValue) bool {
	return i.id == other.id && i.value == other.value
}

func (i *idValue) lessThan(other *idValue) bool {
	return i.id < other.id
}

type mapType string

const (
	defaultMap  = "defaultMap"
	priorityMap = "priorityMap"
)

func createMap[K any, V any, C comparable](mapType mapType, hasher Hasher[K, C],
	equaler Equaler[K], comparator Comparator[K]) Map[K, V] {
	if mapType == defaultMap {
		return NewMap[K, V, C](hasher, equaler)
	} else if mapType == priorityMap {
		return NewPriorityMap[K, V, C](comparator, hasher, equaler)
	}

	panic("Unsupported set type: " + mapType)
}

func testMap(mapType mapType) {
	Describe("can work with basic types.", func() {
		testBasicTypesForMap[int](mapType, func(value int) int {
			return value
		})

		testBasicTypesForMap[float32](mapType, func(value int) float32 {
			return float32(value)
		})

		testBasicTypesForMap[string](mapType, func(value int) string {
			return strconv.Itoa(value)
		})

		testBasicTypesForMap[bool](mapType, func(value int) bool {
			return value != 0
		})
	})

	Describe("can work with other types.", func() {
		var mapForTest Map[*idValue, int]

		BeforeEach(func() {
			mapForTest = createMap[*idValue, int, int](mapType, (*idValue).hash, (*idValue).equals, (*idValue).lessThan)
		})

		It("can contain keys that have the different hash codes.", func() {
			t1 := &idValue{id: 1, value: 1}
			t2 := &idValue{id: 2, value: 2}
			mapForTest.Put(t1, 0)
			mapForTest.Put(t2, 1)

			Expect(mapForTest.ContainsKey(t1)).To(BeTrue())
			value, existing := mapForTest.Get(t1)
			Expect(existing).To(BeTrue())
			Expect(value).To(Equal(0))

			Expect(mapForTest.ContainsKey(t2)).To(BeTrue())
			value, existing = mapForTest.Get(t2)
			Expect(existing).To(BeTrue())
			Expect(value).To(Equal(1))
		})

		It("can contain keys that have the same hash code.", func() {
			t1 := &idValue{id: 1, value: 1}
			t2 := &idValue{id: 1, value: 2}
			mapForTest.Put(t1, 0)
			mapForTest.Put(t2, 1)

			Expect(mapForTest.ContainsKey(t1)).To(BeTrue())
			value, existing := mapForTest.Get(t1)
			Expect(existing).To(BeTrue())
			Expect(value).To(Equal(0))

			Expect(mapForTest.ContainsKey(t2)).To(BeTrue())
			value, existing = mapForTest.Get(t2)
			Expect(existing).To(BeTrue())
			Expect(value).To(Equal(1))
		})

		It("can does not contain keys that have never been put.", func() {
			t1 := &idValue{id: 1, value: 1}
			// different Key but same hash code
			t2 := &idValue{id: 1, value: 2}
			// different Key and different hash code
			t3 := &idValue{id: 2, value: 2}
			mapForTest.Put(t1, 0)

			Expect(mapForTest.ContainsKey(t2)).To(BeFalse())
			Expect(mapForTest.ContainsKey(t3)).To(BeFalse())
		})

		It("can overwrite what it puts.", func() {
			t1 := &idValue{id: 1, value: 1}
			t2 := &idValue{id: 1, value: 1}
			_, existing := mapForTest.Put(t1, 0)
			Expect(existing).To(BeFalse())

			old, existing := mapForTest.Put(t1, 1)
			Expect(existing).To(BeTrue())
			Expect(old).To(Equal(0))

			value, existing := mapForTest.Get(t1)
			Expect(existing).To(BeTrue())
			Expect(value).To(Equal(1))

			old, existing = mapForTest.Put(t2, 2)
			Expect(existing).To(BeTrue())
			Expect(old).To(Equal(1))

			value, existing = mapForTest.Get(t1)
			Expect(existing).To(BeTrue())
			Expect(value).To(Equal(2))
		})

		It("won't panic when trying to remove an nonexistent Key.", func() {
			t1 := &idValue{id: 1, value: 1}
			_, existing := mapForTest.Remove(t1)
			Expect(existing).To(BeFalse())
		})

		It("can remove what it puts when the keys' hash codes are different.", func() {
			t1 := &idValue{id: 1, value: 1}
			t2 := &idValue{id: 2, value: 1}
			mapForTest.Put(t1, 0)
			mapForTest.Put(t2, 1)

			old, existing := mapForTest.Remove(t1)
			Expect(existing).To(BeTrue())
			Expect(old).To(Equal(0))

			value, existing := mapForTest.Remove(t2)
			Expect(existing).To(BeTrue())
			Expect(value).To(Equal(1))
		})

		It("can remove what it puts when the keys have the same hash code.", func() {
			t1 := &idValue{id: 1, value: 1}
			t2 := &idValue{id: 1, value: 2}
			mapForTest.Put(t1, 0)
			mapForTest.Put(t2, 1)

			old, existing := mapForTest.Remove(t1)
			Expect(existing).To(BeTrue())
			Expect(old).To(Equal(0))

			value, existing := mapForTest.Remove(t2)
			Expect(existing).To(BeTrue())
			Expect(value).To(Equal(1))
		})

		It("can return the number of keys it contains.", func() {
			t1 := &idValue{id: 1, value: 1}
			t2 := &idValue{id: 1, value: 1}
			t3 := &idValue{id: 1, value: 2}
			t4 := &idValue{id: 2, value: 2}
			t5 := &idValue{id: 3, value: 2}
			Expect(mapForTest.Len()).To(Equal(0))
			mapForTest.Put(t1, 0)
			Expect(mapForTest.Len()).To(Equal(1))
			mapForTest.Put(t2, 1)
			Expect(mapForTest.Len()).To(Equal(1))
			mapForTest.Put(t3, 1)
			Expect(mapForTest.Len()).To(Equal(2))
			mapForTest.Put(t4, 1)
			Expect(mapForTest.Len()).To(Equal(3))

			mapForTest.Remove(t5)
			Expect(mapForTest.Len()).To(Equal(3))
			mapForTest.Remove(t1)
			Expect(mapForTest.Len()).To(Equal(2))
			mapForTest.Remove(t1)
			Expect(mapForTest.Len()).To(Equal(2))
			mapForTest.Remove(t2)
			Expect(mapForTest.Len()).To(Equal(2))
			mapForTest.Remove(t3)
			Expect(mapForTest.Len()).To(Equal(1))
			mapForTest.Remove(t4)
			Expect(mapForTest.Len()).To(Equal(0))
		})
	})

	Describe("implements Collection interface.", func() {
		var collectionForTest Collection[Pair[int, int]]
		var mapForTest Map[int, int]

		BeforeEach(func() {
			mapForTest = NewMap[int, int, int](basicHasher[int], basicEquator[int])
			collectionForTest = mapForTest
		})

		It("can add Pairs.", func() {
			p := Pair[int, int]{Key: 1, Value: 2}
			collectionForTest.Add(p)
			Expect(collectionForTest.Has(p)).To(BeTrue())
			value, existing := mapForTest.Get(1)
			Expect(existing).To(BeTrue())
			Expect(value).To(Equal(2))

			// overwrite
			p = Pair[int, int]{Key: 1, Value: 1}
			collectionForTest.Add(p)
			Expect(collectionForTest.Has(p)).To(BeTrue())
			value, existing = mapForTest.Get(1)
			Expect(existing).To(BeTrue())
			Expect(value).To(Equal(1))
		})

		It("can work with RemoveFirst.", func() {
			p := Pair[int, int]{Key: 1, Value: 2}
			collectionForTest.Add(p)
			collectionForTest.RemoveFirst(p)
			Expect(collectionForTest.Has(p)).To(BeFalse())

			// Won't panic
			collectionForTest.RemoveFirst(p)
		})

		It("can pop items.", func() {
			p := Pair[int, int]{Key: 1, Value: 2}
			collectionForTest.Add(p)
			pair, existing := collectionForTest.Pop()
			Expect(existing).To(BeTrue())
			Expect(pair.Key).To(Equal(1))
			Expect(pair.Value).To(Equal(2))
			pair, existing = collectionForTest.Pop()
			Expect(existing).To(BeFalse())

			// Work with multiple items
			collectionForTest.Add(p)
			p = Pair[int, int]{Key: 2, Value: 2}
			collectionForTest.Add(p)
			pair, existing = collectionForTest.Pop()
			Expect(existing).To(BeTrue())
			Expect(pair.Key).To(Or(Equal(1), Equal(2)))
			Expect(pair.Value).To(Equal(2))
			pair, existing = collectionForTest.Pop()
			Expect(existing).To(BeTrue())
			Expect(pair.Key).To(Or(Equal(1), Equal(2)))
			Expect(pair.Value).To(Equal(2))
			pair, existing = collectionForTest.Pop()
			Expect(existing).To(BeFalse())
		})

		It("can return the number of items it contains.", func() {
			p := Pair[int, int]{Key: 1, Value: 2}
			collectionForTest.Add(p)
			Expect(collectionForTest.Len()).To(Equal(1))

			p = Pair[int, int]{Key: 1, Value: 3}
			collectionForTest.Add(p)
			Expect(collectionForTest.Len()).To(Equal(1))

			p = Pair[int, int]{Key: 2, Value: 3}
			collectionForTest.Add(p)
			Expect(collectionForTest.Len()).To(Equal(2))

			collectionForTest.RemoveFirst(p)
			Expect(collectionForTest.Len()).To(Equal(1))

			collectionForTest.RemoveFirst(p)
			Expect(collectionForTest.Len()).To(Equal(1))

			collectionForTest.Pop()
			Expect(collectionForTest.Len()).To(Equal(0))
		})
	})
}

var _ = Describe("DefaultMap", func() {
	testMap(defaultMap)
})
