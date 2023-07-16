package gochanbroker

import (
	"sync"
)

// Thread safe list implementation

type safeListStruct[T comparable] struct {
	arr   *[]T
	mutex *sync.Mutex
}

func MakeSafeList[T comparable](bufflen int) safeListStruct[T] {
	arr := make([]T, 0, bufflen)
	return safeListStruct[T]{
		arr:   &arr,
		mutex: &sync.Mutex{},
	}
}

func (q *safeListStruct[T]) Duplicate() safeListStruct[T] {
	newArr := make([]T, len(*q.arr))
	copy(newArr, *q.arr)

	return safeListStruct[T]{
		arr:   &newArr,
		mutex: &sync.Mutex{},
	}
}

func (q *safeListStruct[T]) Len() int {
	return len(*q.arr)
}

func (q *safeListStruct[T]) Push(item T) {
	q.mutex.Lock()
	defer q.mutex.Unlock()

	newArr := append(*q.arr, item)
	q.arr = &newArr
}

func (q *safeListStruct[T]) First() (T, bool) {
	q.mutex.Lock()
	defer q.mutex.Unlock()

	l := q.Len()

	if l > 0 {
		return (*q.arr)[0], true
	}

	var empty T
	return empty, false
}

func (q *safeListStruct[T]) Remove(item T) bool {
	q.mutex.Lock()
	defer q.mutex.Unlock()

	for i, v := range *q.arr {
		if v == item {
			newArr := append((*q.arr)[:i], (*q.arr)[i+1:]...)
			q.arr = &newArr
			return true
		}
	}
	return false
}

func (q *safeListStruct[T]) Pop() (T, bool) {
	q.mutex.Lock()
	defer q.mutex.Unlock()

	l := q.Len()
	if l > 0 {
		v := (*q.arr)[l-1]
		slice := (*q.arr)[0 : l-1]
		q.arr = &slice
		return v, true
	}
	var empty T

	return empty, false
}
