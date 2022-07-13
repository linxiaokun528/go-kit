package collection_test

import (
	"fmt"
	"reflect"
	"sort"
	"strconv"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	. "go-kit/pkg/util/collection"
)

func testBasicTypesForSet[T comparable](setType setType, convert fromInt[T]) {
	var obj T
	typeName := reflect.TypeOf(obj).Name()

	Describe(fmt.Sprintf("It can work with type [%s].", typeName), func() {
		var setForTest Set[T]

		BeforeEach(func() {
			// use a fake comparator here. We don't really need a comparator in this test
			setForTest = createSet[T, T](setType, basicHasher[T], basicEquator[T], fakeComparator[T])
		})

		It("can check if an item is existent.", func() {
			Expect(setForTest.Has(convert(0))).To(BeFalse())

			setForTest.Add(convert(0))
			Expect(setForTest.Has(convert(0))).To(BeTrue())
			setForTest.Add(convert(1))
			Expect(setForTest.Has(convert(1))).To(BeTrue())
		})

		It("can remove what it adds.", func() {
			Expect(setForTest.RemoveFirst(convert(0))).To(BeFalse())

			setForTest.Add(convert(0))
			Expect(setForTest.RemoveFirst(convert(0))).To(BeTrue())
			Expect(setForTest.RemoveFirst(convert(1))).To(BeFalse())
		})

		It("can return what it adds into an array.", func() {
			Expect(setForTest.ToArray()).To(BeEmpty())

			setForTest.Add(convert(0))
			setForTest.Add(convert(1))
			Expect(setForTest.ToArray()).To(ConsistOf(convert(0), convert(1)))

			setForTest.RemoveFirst(convert(0))
			Expect(setForTest.ToArray()).To(ConsistOf(convert(1)))
			setForTest.RemoveFirst(convert(1))
			Expect(setForTest.ToArray()).To(BeEmpty())
			setForTest.RemoveFirst(convert(1))
			Expect(setForTest.ToArray()).To(BeEmpty())
		})

		It("can try to pop items.", func() {
			_, exists := setForTest.TryPop()
			Expect(exists).To(BeFalse())

			setForTest.Add(convert(0))
			value, exists := setForTest.TryPop()
			Expect(exists).To(BeTrue())
			Expect(value).To(Equal(convert(0)))

			// Work with multiple items
			setForTest.Add(convert(0))
			setForTest.Add(convert(1))
			value, exists = setForTest.TryPop()
			Expect(exists).To(BeTrue())
			Expect(value).To(Or(Equal(convert(1)), Equal(convert(0))))
			value, exists = setForTest.TryPop()
			Expect(exists).To(BeTrue())
			Expect(value).To(Or(Equal(convert(1)), Equal(convert(0))))
			value, exists = setForTest.TryPop()
			Expect(exists).To(BeFalse())
		})

		It("can return the number of items it contains.", func() {
			Expect(setForTest.Len()).To(Equal(0))
			setForTest.Add(convert(0))
			Expect(setForTest.Len()).To(Equal(1))
			setForTest.RemoveFirst(convert(1))
			Expect(setForTest.Len()).To(Equal(1))
			setForTest.RemoveFirst(convert(0))
			Expect(setForTest.Len()).To(Equal(0))

			setForTest.Add(convert(1))
			Expect(setForTest.Len()).To(Equal(1))
			setForTest.Add(convert(0))
			Expect(setForTest.Len()).To(Equal(2))
			setForTest.TryPop()
			Expect(setForTest.Len()).To(Equal(1))
			setForTest.TryPop()
			Expect(setForTest.Len()).To(Equal(0))
			setForTest.TryPop()
			Expect(setForTest.Len()).To(Equal(0))
		})

		It("can clear what it puts.", func() {
			setForTest.Add(convert(0))
			setForTest.Add(convert(1))
			setForTest.Clear()
			Expect(setForTest.Len()).To(Equal(0))

			setForTest.Add(convert(0))
			Expect(setForTest.Has(convert(0))).To(BeTrue())
			Expect(setForTest.Has(convert(1))).To(BeFalse())
		})

	})
}

type setType string

const (
	defaultSet    = "defaultSet"
	prioritySet   = "prioritySet"
	threadSafeSet = "threadSafeSet"
)

func createSet[T any, C comparable](setType setType, hasher Hasher[T, C],
	equaler Equaler[T], comparator Comparator[T]) Set[T] {
	if setType == defaultSet {
		return NewSet[T, C](hasher, equaler)
	} else if setType == prioritySet {
		return NewPrioritySet[T, C](comparator, hasher, equaler)
	} else if setType == threadSafeSet {
		return NewThreadSafeSet[T, C](hasher, equaler)
	}

	panic("Unsupported set type: " + setType)
}

// I know it's ugly, but you can't pass a parameter like `create func[T any]() Set[T]`, which will have an error
// "Function type cannot have type parameters"
func testSet(setType setType) {
	Describe("can work with basic types.", func() {
		testBasicTypesForSet[int](setType, func(value int) int {
			return value
		})

		testBasicTypesForSet[float32](setType, func(value int) float32 {
			return float32(value)
		})

		testBasicTypesForSet[string](setType, func(value int) string {
			return strconv.Itoa(value)
		})

		testBasicTypesForSet[bool](setType, func(value int) bool {
			return value != 0
		})
	})

	Describe("can work with other types.", func() {
		var setForTest Set[*idValue]

		BeforeEach(func() {
			setForTest = createSet[*idValue, int](setType, (*idValue).hash, (*idValue).equals, (*idValue).lessThan)
		})

		It("can contain items that have the different hash codes.", func() {
			t1 := &idValue{id: 1, value: 1}
			t2 := &idValue{id: 2, value: 2}

			Expect(setForTest.Has(t1)).To(BeFalse())
			Expect(setForTest.Has(t2)).To(BeFalse())
			setForTest.Add(t1)
			setForTest.Add(t2)
			Expect(setForTest.Has(t1)).To(BeTrue())
			Expect(setForTest.Has(t2)).To(BeTrue())
		})

		It("can contain items that have the same hash code.", func() {
			t1 := &idValue{id: 1, value: 1}
			t2 := &idValue{id: 1, value: 2}

			Expect(setForTest.Has(t1)).To(BeFalse())
			Expect(setForTest.Has(t2)).To(BeFalse())
			setForTest.Add(t1)
			setForTest.Add(t2)
			Expect(setForTest.Has(t1)).To(BeTrue())
			Expect(setForTest.Has(t2)).To(BeTrue())
		})

		It("can remove what it puts when the keys' hash codes are different.", func() {
			t1 := &idValue{id: 1, value: 1}
			t2 := &idValue{id: 2, value: 1}
			setForTest.Add(t1)
			setForTest.Add(t2)

			Expect(setForTest.RemoveFirst(t1)).To(BeTrue())
			Expect(setForTest.Len()).To(Equal(1))

			Expect(setForTest.RemoveFirst(t1)).To(BeFalse())
			Expect(setForTest.Len()).To(Equal(1))

			Expect(setForTest.RemoveFirst(t2)).To(BeTrue())
			Expect(setForTest.Len()).To(Equal(0))
		})

		It("can remove what it puts when the keys have the same hash code.", func() {
			t1 := &idValue{id: 1, value: 1}
			t2 := &idValue{id: 1, value: 2}
			setForTest.Add(t1)
			setForTest.Add(t2)

			Expect(setForTest.RemoveFirst(t1)).To(BeTrue())
			Expect(setForTest.Len()).To(Equal(1))

			Expect(setForTest.RemoveFirst(t1)).To(BeFalse())
			Expect(setForTest.Len()).To(Equal(1))

			Expect(setForTest.RemoveFirst(t2)).To(BeTrue())
			Expect(setForTest.Len()).To(Equal(0))
		})

		It("can return the number of keys it contains.", func() {
			t1 := &idValue{id: 1, value: 1}
			t2 := &idValue{id: 1, value: 1}
			t3 := &idValue{id: 1, value: 2}
			t4 := &idValue{id: 2, value: 2}
			t5 := &idValue{id: 3, value: 2}
			Expect(setForTest.Len()).To(Equal(0))
			setForTest.Add(t1)
			Expect(setForTest.Len()).To(Equal(1))
			setForTest.Add(t2)
			Expect(setForTest.Len()).To(Equal(1))
			setForTest.Add(t3)
			Expect(setForTest.Len()).To(Equal(2))
			setForTest.Add(t4)
			Expect(setForTest.Len()).To(Equal(3))

			setForTest.RemoveFirst(t5)
			Expect(setForTest.Len()).To(Equal(3))
			setForTest.RemoveFirst(t1)
			Expect(setForTest.Len()).To(Equal(2))
			setForTest.RemoveFirst(t1)
			Expect(setForTest.Len()).To(Equal(2))
			setForTest.RemoveFirst(t2)
			Expect(setForTest.Len()).To(Equal(2))
			setForTest.RemoveFirst(t3)
			Expect(setForTest.Len()).To(Equal(1))
			setForTest.RemoveFirst(t4)
			Expect(setForTest.Len()).To(Equal(0))
		})
	})
}

var _ = Describe("Default set", func() {
	testSet(defaultSet)
})

var _ = Describe("ThreadSafeSet", func() {
	testSet(threadSafeSet)

	var concurrentLevel int
	var setForTest Set[int]

	BeforeEach(func() {
		concurrentLevel = 30
		setForTest = NewThreadSafeSet[int, int](basicHasher[int], basicEquator[int])
	})

	It("can add items concurrently", func() {
		for i := 0; i < concurrentLevel; i++ {
			tmp := i
			go func() {
				setForTest.Add(tmp)
			}()
		}

		Eventually(setForTest.Len).Should(Equal(concurrentLevel))

		Expect(setForTest.Len()).To(Equal(concurrentLevel))
		data := setForTest.ToArray()
		sort.Ints(data)
		for i, datum := range data {
			Expect(datum).To(Equal(i))
		}
	})

	It("can pop items concurrently", func() {
		for i := 0; i < concurrentLevel; i++ {
			setForTest.Add(i)
		}

		for i := 0; i < concurrentLevel; i++ {
			go setForTest.TryPop()
		}

		Eventually(setForTest.Len).Should(Equal(0))
	})
})
