package utils

import "sync"

type queueItem[T any] struct {
	value any
	next  *queueItem[T]
}

type Queue[T any] struct {
	head          *queueItem[T]
	tail          *queueItem[T]
	cond          *sync.Cond
	lock          *sync.Mutex
	deBlock       bool
	ignoreEnqueue bool
}

func NewQueue[T any]() *Queue[T] {
	lock := &sync.Mutex{}
	return &Queue[T]{
		head: nil,
		tail: nil,
		cond: sync.NewCond(lock),
		lock: lock,
	}
}

func (q *Queue[T]) Enqueue(value T) {
	q.lock.Lock()
	defer q.lock.Unlock()
	if q.ignoreEnqueue {
		return
	}
	if q.head == nil {
		q.head = &queueItem[T]{value: value}
		q.tail = q.head
	} else {
		q.tail.next = &queueItem[T]{value: value}
		q.tail = q.tail.next
	}
	q.cond.Signal()
}

func (q *Queue[T]) Dequeue() T {
	q.lock.Lock()
	defer q.lock.Unlock()
	for q.head == nil && !q.deBlock {
		q.cond.Wait()
	}
	var value T
	if q.head == nil {
		return value
	}
	var dValue T = value
	value = q.head.value
	if q.head == q.tail {
		q.tail = nil
	}
	oHead := q.head
	q.head = q.head.next
	oHead.value = dValue
	oHead.next = nil
	return value
}

func (q *Queue[T]) StartUnBlocking() {
	q.lock.Lock()
	defer q.lock.Unlock()
	q.deBlock = true
	q.ignoreEnqueue = true
	q.cond.Broadcast()
}

func (q *Queue[T]) EndUnBlocking() {
	q.lock.Lock()
	defer q.lock.Unlock()
	q.deBlock = false
	q.ignoreEnqueue = false
}

func (q *Queue[T]) IsUnBlocking() bool {
	q.lock.Lock()
	defer q.lock.Unlock()
	return q.deBlock
}

func (q *Queue[T]) Clear() {
	q.lock.Lock()
	defer q.lock.Unlock()
	var dValue T
	cHead := q.head
	for cHead != nil {
		oHead := cHead
		cHead = oHead.next
		oHead.value = dValue
		oHead.next = nil
	}
}
