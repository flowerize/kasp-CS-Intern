package queue

import (
	"errors"
	"kasp-cs-tsvetkov-denis/internal/worker"
	"sync"
)

type TaskQueueInterface interface {
	Enqueue(task worker.Task) error
	Dequeue() (worker.Task, error)
	Close() error
	Size() int
}

// Ошибки очереди
var (
	ErrQueueFull   = errors.New("queue is full")
	ErrQueueEmpty  = errors.New("queue is empty")
	ErrQueueClosed = errors.New("queue is closed")
)

// Queue представляет буферизированную очередь с фиксированным capacity.
type Queue[T worker.Task] struct {
	mu       sync.Mutex
	notEmpty *sync.Cond
	notFull  *sync.Cond

	buffer   []T
	head     int
	tail     int
	size     int
	capacity int
	closed   bool
}

// New создает новую очередь с указанным capacity.
func New[T worker.Task](size int) *Queue[T] {
	q := &Queue[T]{
		buffer:   make([]T, size),
		capacity: size,
	}
	q.notEmpty = sync.NewCond(&q.mu)
	q.notFull = sync.NewCond(&q.mu)
	return q
}

// Enqueue добавляет элемент в очередь.
// Возвращает ошибку, если очередь закрыта или полна.
func (q *Queue[T]) Enqueue(item T) error {
	q.mu.Lock()
	defer q.mu.Unlock()

	if q.closed {
		return ErrQueueClosed
	}
	if q.size == q.capacity {
		return ErrQueueFull
	}
	q.buffer[q.tail] = item
	q.tail = (q.tail + 1) % q.capacity
	q.size++
	q.notEmpty.Signal()
	return nil
}

// Dequeue забирает элемент из очереди.
// Возвращает ошибку, если очередь пуста или закрыта.
func (q *Queue[T]) Dequeue() (worker.Task, error) {
	var zero T

	q.mu.Lock()
	defer q.mu.Unlock()

	for q.size == 0 && !q.closed {
		q.notEmpty.Wait()
	}

	if q.size == 0 && q.closed {
		return zero, ErrQueueClosed
	}
	item := q.buffer[q.head]
	q.head = (q.head + 1) % q.capacity
	q.size--
	q.notFull.Signal()
	return item, nil
}

// Close закрывает очередь. После закрытия нельзя добавлять элементы.
// Все ожидающие операции разблокируются с ошибкой ErrQueueClosed.
func (q *Queue[T]) Close() error {
	q.mu.Lock()
	defer q.mu.Unlock()
	if q.closed {
		return errors.New("queue already closed")
	}
	q.closed = true
	q.notEmpty.Broadcast()
	q.notFull.Broadcast()
	return nil
}

// Size возвращает текущее количество элементов в очереди.
func (q *Queue[T]) Size() int {
	q.mu.Lock()
	defer q.mu.Unlock()
	return q.size
}
