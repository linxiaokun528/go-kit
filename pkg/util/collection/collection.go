package collection

// Interface is describing a collection.
type Collection[T any] interface {
	Add(item T)
	RemoveFirst(item T) bool
	Pop() (T, bool)
	Has(item T) bool
	Len() int
}

//type EnhancedCollection[T any] struct {
//	*Collection[T]
//}
//
//// Will add the following methods when needed
//func (e *EnhancedCollection[T]) AddAll(items ...T) {
//
//}
//
//func (e *EnhancedCollection[T]) RemoveAll(item T) {
//
//}
//
//func (e *EnhancedCollection[T]) Copy() *EnhancedCollection[T] {
//
//}
//
//func (e *EnhancedCollection[T]) IsSubset(c Collection[T]) bool {
//
//}
//
//func (e *EnhancedCollection[T]) IsSuperset(c Collection[T]) bool {
//
//}
//
//func (e *EnhancedCollection[T]) Filter(func(T) bool) {
//
//}
//
//func (e *EnhancedCollection[T]) List() []T {
//
//}
//
//func (e *EnhancedCollection[T]) Concatenate(c Collection[T]) EnhancedCollection[T] {
//
//}
//
//func (e *EnhancedCollection[T]) IsEmpty() bool {
//
//}
//
//func (e *EnhancedCollection[T]) Clear() bool {
//
//}
