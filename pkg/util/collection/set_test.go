package collection_test

import (
	"fmt"
	"reflect"
	"strconv"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go-kit/pkg/util/collection"
)

func testBasicTypesForSet[T comparable](convert fromInt[T]) {
	var obj T
	typeName := reflect.TypeOf(obj).Name()

	Describe(fmt.Sprintf("It can work with type [%s].", typeName), func() {
		var setForTest collection.Collection[T]

		BeforeEach(func() {
			setForTest = collection.NewSet[T](basicHasher[T], basicEquator[T])
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

		It("can pop items.", func() {
			_, existing := setForTest.Pop()
			Expect(existing).To(BeFalse())

			setForTest.Add(convert(0))
			value, existing := setForTest.Pop()
			Expect(existing).To(BeTrue())
			Expect(value).To(Equal(convert(0)))

			// Work with multiple items
			setForTest.Add(convert(0))
			setForTest.Add(convert(1))
			value, existing = setForTest.Pop()
			Expect(existing).To(BeTrue())
			Expect(value).To(Or(Equal(convert(1)), Equal(convert(0))))
			value, existing = setForTest.Pop()
			Expect(existing).To(BeTrue())
			Expect(value).To(Or(Equal(convert(1)), Equal(convert(0))))
			value, existing = setForTest.Pop()
			Expect(existing).To(BeFalse())
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
			setForTest.Pop()
			Expect(setForTest.Len()).To(Equal(1))
			setForTest.Pop()
			Expect(setForTest.Len()).To(Equal(0))
			setForTest.Pop()
			Expect(setForTest.Len()).To(Equal(0))
		})
	})
}

var _ = Describe("Set", func() {
	Describe("can work with basic types.", func() {
		testBasicTypesForSet[int](func(value int) int {
			return value
		})

		testBasicTypesForSet[float32](func(value int) float32 {
			return float32(value)
		})

		testBasicTypesForSet[string](func(value int) string {
			return strconv.Itoa(value)
		})

		testBasicTypesForSet[bool](func(value int) bool {
			return value != 0
		})
	})

	Describe("can work with other types.", func() {
		var setForTest collection.Collection[*idValue]

		BeforeEach(func() {
			setForTest = collection.NewSet[*idValue]((*idValue).hash, (*idValue).equals)
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
})
