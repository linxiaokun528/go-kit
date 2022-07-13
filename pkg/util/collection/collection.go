package collection

// Collection To avoid Value copy, you may want T to be pointer types.
//  However, if T is a pointer type, we must make sure that the hash code remains the same.
type Collection[T any] interface {
	// Add for some collections like set/map, we need to return the replaced item
	Add(item T) (oldItem T, replaced bool)
	RemoveFirst(item T) bool
	TryPop() (T, bool)
	Has(item T) bool
	Len() int
	Clear()
	ToArray() []T // The order will not be guaranteed
}

//type CollectionTool[T any] struct {
//}
//
//// Will add the following methods when needed
//
//func (c *CollectionTool[T]) Pop(collection Collection[T]) T {
//    // If the collection is empty, panic
//}
//
//func (c *CollectionTool[T]) AddAll(collection Collection[T], items ...T) {
//
//}
//
//func (c *CollectionTool[T]) Extend(c1, c2 Collection[T]) {
//
//}
//
//func (c *CollectionTool[T]) RemoveAll(collection Collection[T], item T) {
//
//}
//
//func (c *CollectionTool[T]) IsSubset(c1, c2 Collection[T]) bool {
//
//}
//
//func (c *CollectionTool[T]) IsSuperset(c1, c2 Collection[T]) bool {
//
//}
//
//func (c *CollectionTool[T]) Filter(collection Collection[T], func(T) bool) Collection[T] {
//
//}
//
//
//func (c *CollectionTool[T]) Concatenate(c1, c2 Collection[T]) CollectionTool[T] {
//
//}
//
//func (c *CollectionTool[T]) IsEmpty(collection Collection[T]) bool {
//
//}
