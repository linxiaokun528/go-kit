package collection_test

import (
	"fmt"
	"math/rand"
	"sort"
	"strconv"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	. "go-kit/pkg/util/collection"
)

func intAscComparator(first, second int) bool {
	return first <= second
}

func intDescComparator(first, second int) bool {
	return first >= second
}

func getRandomArray(num int) (result []int) {
	for i := 0; i < num; i++ {
		result = append(result, rand.Intn(num))
	}
	return
}

func getSequence(num int) (result []int) {
	for i := 0; i < num; i++ {
		result = append(result, i)
	}
	return
}

func permutation(data []int) [][]int {
	return permutation_recursive(data, 0)
}

func permutation_recursive(data []int, start int) (result [][]int) {
	if start == len(data)-1 {
		tmp := make([]int, len(data))
		copy(tmp, data)
		result = append(result, tmp)
		return
	}

	for i := start; i < len(data); i++ {
		data[start], data[i] = data[i], data[start]
		result = append(result, permutation_recursive(data, start+1)...)
		data[start], data[i] = data[i], data[start]
	}
	return
}

func intUniquer(array []int) []int {
	m := map[int]int{}
	for _, data := range array {
		m[data] = 0
	}
	result := []int{}
	for k := range m {
		result = append(result, k)
	}

	return result
}

func pairUniquer(array []Pair[int, int]) []Pair[int, int] {
	m := map[int]Pair[int, int]{}
	for _, data := range array {
		m[data.Key] = data
	}
	result := []Pair[int, int]{}
	for _, pair := range m {
		result = append(result, pair)
	}

	return result
}

func fakeUniquer[T any](array []T) []T {
	return array
}

func intComparator(first, second int) bool {
	return first <= second
}

func pairComparator(first, second Pair[int, int]) bool {
	return first.Key <= second.Key
}

func intToPair(array []int) []Pair[int, int] {
	var pairs []Pair[int, int]
	for _, entry := range array {
		pairs = append(pairs, Pair[int, int]{Key: entry, Value: entry + 1})
	}
	return pairs
}

func fakeHasher(int) int {
	return 0
}

func testCollection[T any](c PriorityCollection[T], array []T, comparator Comparator[T],
	ascending bool, uniquer func([]T) []T) {
	c.Clear()

	for _, elem := range array {
		c.Add(elem)
	}

	expected := make([]T, len(array))
	copy(expected, array)
	// remove the repetitive entries array `expected`
	expected = uniquer(expected)
	// sort the array
	sort.Slice(expected, func(i, j int) bool {
		less := comparator(expected[i], expected[j])
		if ascending {
			return less
		}

		return !less
	})

	actual := []T{}
	for value, existing := c.Pop(); existing; value, existing = c.Pop() {
		actual = append(actual, value)
	}

	test_type := "ascending"
	if !ascending {
		test_type = "descending"
	}

	Expect(actual).To(Equal(expected), fmt.Sprintf("Case failed in %s test: %v", test_type, array))
}

var _ = BeforeSuite(func() {
	rand.Seed(time.Now().UnixNano())
})

var _ = Describe("PriorityCollection should be tested multiple times.", func() {
	var testTimesForRamdomCase int
	var arrayLengths []int
	var maxLengthForPermutation int

	BeforeEach(func() {
		testTimesForRamdomCase = 30
		arrayLengths = []int{0, 1, 2, 3, 4, 8, 10, 30}
		maxLengthForPermutation = 9
	})

	Describe("PriorityQueue", func() {
		Describe("can pop the element in order.", func() {
			It("works with sequence.", func() {
				for _, length := range arrayLengths {
					// It will take just too long.
					if length > maxLengthForPermutation {
						continue
					}

					for _, data := range permutation(getSequence(length)) {
						testCollection[int](NewPriorityQueue[int](intAscComparator, basicEquator[int]),
							data, intComparator, true, fakeUniquer[int])
						testCollection[int](NewPriorityQueue[int](intDescComparator, basicEquator[int]),
							data, intComparator, false, fakeUniquer[int])
					}

				}
			})

			It("works with arrays with repetition values.", func() {
				for i := 0; i < testTimesForRamdomCase; i++ {
					for _, length := range arrayLengths {
						testCollection[int](NewPriorityQueue[int](intAscComparator, basicEquator[int]),
							getRandomArray(length), intComparator, true, fakeUniquer[int])
						testCollection[int](NewPriorityQueue[int](intDescComparator, basicEquator[int]),
							getRandomArray(length), intComparator, false, fakeUniquer[int])
					}
				}
			})
		})

		Describe("It supports Collection interface.", func() {
			var priorityQueue PriorityQueue[int]

			BeforeEach(func() {
				priorityQueue = NewPriorityQueue(intAscComparator, basicEquator[int])
			})

			It("can add items.", func() {
				_, replaced := priorityQueue.Add(1)
				Expect(replaced).To(BeFalse())

				_, replaced = priorityQueue.Add(1)
				Expect(replaced).To(BeFalse())

				_, replaced = priorityQueue.Add(2)
				Expect(replaced).To(BeFalse())
			})

			It("can remove the item it adds.", func() {
				Expect(priorityQueue.RemoveFirst(0)).To(BeFalse())
				priorityQueue.Add(1)
				Expect(priorityQueue.RemoveFirst(0)).To(BeFalse())
				Expect(priorityQueue.RemoveFirst(1)).To(BeTrue())
				Expect(priorityQueue.RemoveFirst(1)).To(BeFalse())
				priorityQueue.Add(1)
				priorityQueue.Add(1)
				Expect(priorityQueue.RemoveFirst(1)).To(BeTrue())
				Expect(priorityQueue.RemoveFirst(1)).To(BeTrue())
				Expect(priorityQueue.RemoveFirst(1)).To(BeFalse())
			})

			It("can check if an item exists.", func() {
				Expect(priorityQueue.Has(0)).To(BeFalse())
				priorityQueue.Add(1)
				Expect(priorityQueue.Has(0)).To(BeFalse())
				Expect(priorityQueue.Has(1)).To(BeTrue())
				priorityQueue.RemoveFirst(1)
				Expect(priorityQueue.Has(1)).To(BeFalse())

				priorityQueue.Add(1)
				priorityQueue.Add(1)
				Expect(priorityQueue.Has(1)).To(BeTrue())
				priorityQueue.RemoveFirst(1)
				Expect(priorityQueue.Has(1)).To(BeTrue())
				priorityQueue.RemoveFirst(1)
				Expect(priorityQueue.Has(1)).To(BeFalse())
			})

			It("can return the number of the items it contains.", func() {
				Expect(priorityQueue.Len()).To(Equal(0))
				priorityQueue.Add(1)
				Expect(priorityQueue.Len()).To(Equal(1))
				priorityQueue.Add(1)
				Expect(priorityQueue.Len()).To(Equal(2))
				priorityQueue.Add(2)
				Expect(priorityQueue.Len()).To(Equal(3))
				priorityQueue.RemoveFirst(1)
				Expect(priorityQueue.Len()).To(Equal(2))
				priorityQueue.RemoveFirst(1)
				Expect(priorityQueue.Len()).To(Equal(1))
				priorityQueue.RemoveFirst(1)
				Expect(priorityQueue.Len()).To(Equal(1))
				priorityQueue.RemoveFirst(2)
				Expect(priorityQueue.Len()).To(Equal(0))
			})

			It("can clear what it puts.", func() {
				priorityQueue.Clear()
				Expect(priorityQueue.Len()).To(Equal(0))

				priorityQueue.Add(1)
				priorityQueue.Clear()
				Expect(priorityQueue.Len()).To(Equal(0))

				priorityQueue.Add(1)
				priorityQueue.Add(2)
				priorityQueue.Clear()
				Expect(priorityQueue.Len()).To(Equal(0))
			})
		})
	})

	Describe("PrioritySet", func() {
		Describe("can pop the element in order.", func() {
			It("works with sequence.", func() {
				for _, length := range arrayLengths {
					// It will take just too long.
					if length > maxLengthForPermutation {
						continue
					}

					for _, data := range permutation(getSequence(length)) {
						testCollection[int](
							NewPrioritySet[int, int](intAscComparator, basicHasher[int], basicEquator[int]),
							data, intComparator, true, intUniquer)
						testCollection[int](
							NewPrioritySet[int, int](intDescComparator, basicHasher[int], basicEquator[int]),
							data, intComparator, false, intUniquer)
					}
				}
			})

			It("works with arrays with repetition values.", func() {
				for i := 0; i < testTimesForRamdomCase; i++ {
					for _, length := range arrayLengths {
						testCollection[int](
							NewPrioritySet[int, int](intAscComparator, basicHasher[int], basicEquator[int]),
							getRandomArray(length), intComparator, true, intUniquer)
						testCollection[int](
							NewPrioritySet[int, int](intDescComparator, basicHasher[int], basicEquator[int]),
							getRandomArray(length), intComparator, false, intUniquer)
					}
				}
			})

			It("works even all the values have the same hash code.", func() {
				for i := 0; i < testTimesForRamdomCase; i++ {
					for _, length := range arrayLengths {
						testCollection[int](
							NewPrioritySet[int, int](intAscComparator, fakeHasher, basicEquator[int]),
							getRandomArray(length), intComparator, true, intUniquer)
						testCollection[int](
							NewPrioritySet[int, int](intDescComparator, fakeHasher, basicEquator[int]),
							getRandomArray(length), intComparator, false, intUniquer)
					}
				}
			})

			Describe("It supports Set interface.", func() {
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
		})
	})

	Describe("PriorityMap", func() {
		Describe("can pop the element in order.", func() {
			It("works with sequence.", func() {
				for _, length := range arrayLengths {
					// It will take just too long.
					if length > maxLengthForPermutation {
						continue
					}

					for _, data := range permutation(getSequence(length)) {
						pairs := intToPair(data)
						testCollection[Pair[int, int]](
							NewPriorityMap[int, int, int](intAscComparator, basicHasher[int], basicEquator[int]),
							pairs, pairComparator, true, pairUniquer)
						testCollection[Pair[int, int]](
							NewPriorityMap[int, int, int](intDescComparator, basicHasher[int], basicEquator[int]),
							pairs, pairComparator, false, pairUniquer)
					}
				}
			})

			It("works with arrays with repetition values.", func() {
				for i := 0; i < testTimesForRamdomCase; i++ {
					for _, length := range arrayLengths {
						testCollection[Pair[int, int]](
							NewPriorityMap[int, int, int](intAscComparator, basicHasher[int], basicEquator[int]),
							intToPair(getRandomArray(length)), pairComparator, true, pairUniquer)
						testCollection[Pair[int, int]](
							NewPriorityMap[int, int, int](intDescComparator, basicHasher[int], basicEquator[int]),
							intToPair(getRandomArray(length)), pairComparator, false, pairUniquer)
					}
				}
			})

			It("works even all the values have the same hash code.", func() {
				for i := 0; i < testTimesForRamdomCase; i++ {
					for _, length := range arrayLengths {
						testCollection[Pair[int, int]](
							NewPriorityMap[int, int, int](intAscComparator, fakeHasher, basicEquator[int]),
							intToPair(getRandomArray(length)), pairComparator, true, pairUniquer)
						testCollection[Pair[int, int]](
							NewPriorityMap[int, int, int](intDescComparator, fakeHasher, basicEquator[int]),
							intToPair(getRandomArray(length)), pairComparator, false, pairUniquer)
					}
				}
			})

			Describe("It supports Map interface.", func() {
				testBasicTypesForMap[int](func(value int) int {
					return value
				})

				testBasicTypesForMap[float32](func(value int) float32 {
					return float32(value)
				})

				testBasicTypesForMap[string](func(value int) string {
					return strconv.Itoa(value)
				})

				testBasicTypesForMap[bool](func(value int) bool {
					return value != 0
				})
			})
		})
	})
})