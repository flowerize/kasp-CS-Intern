package queue_test

import (
	"testing"

	"kasp-cs-tsvetkov-denis/internal/queue"
	"kasp-cs-tsvetkov-denis/internal/worker"

	"github.com/stretchr/testify/assert"
)

type DummyTask struct {
	id string
}

func (t *DummyTask) ID() string {
	return t.id
}

func (t *DummyTask) Execute() error {
	return nil
}

func TestQueue_EnqueueDequeue(t *testing.T) {
	q := queue.New[worker.Task](10)

	task := &DummyTask{id: "test1"}
	err := q.Enqueue(task)
	assert.NoError(t, err)

	dequeuedTask, err := q.Dequeue()
	assert.NoError(t, err)
	assert.Equal(t, task.ID(), dequeuedTask.ID())
}

func TestQueue_Close(t *testing.T) {
	q := queue.New[worker.Task](10)

	assert.NoError(t, q.Close())

	// Попытка положить задачу после закрытия
	err := q.Enqueue(&DummyTask{id: "test2"})
	assert.Error(t, err)

	// Попытка получить задачу из закрытой и пустой очереди
	_, err = q.Dequeue()
	assert.Error(t, err)
}

func TestQueue_Size(t *testing.T) {
	q := queue.New[worker.Task](10)
	assert.Equal(t, 0, q.Size())

	q.Enqueue(&DummyTask{id: "test3"})
	assert.Equal(t, 1, q.Size())

	_, _ = q.Dequeue()
	assert.Equal(t, 0, q.Size())
}
