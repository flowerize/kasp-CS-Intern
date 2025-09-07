package worker

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestWorkerPool_ProcessTaskSuccess(t *testing.T) {
	wp := NewWorkerPool(1, 10)
	wp.Start()
	defer wp.Stop()

	task := NewSimpleTask("task-1", "payload", 2)

	err := wp.Enqueue(task)
	assert.NoError(t, err)

	// Подождать, пока состояние не станет "done" или таймаут
	timeout := time.After(2 * time.Second)
	ticker := time.Tick(10 * time.Millisecond)

	for {
		select {
		case <-timeout:
			t.Fatal("Timeout waiting for task to complete")
		case <-ticker:
			state, ok := wp.GetTaskState(task.ID())
			if ok && state == "done" {
				// успех
				return
			}
		}
	}
}

func TestWorkerPool_CancelContextStopsWorkers(t *testing.T) {
	wp := NewWorkerPool(2, 10)
	wp.Start()

	canceledCtx, cancel := context.WithCancel(context.Background())

	taskRunning := &SimpleTaskRunning{
		id:     "task-running",
		ctx:    canceledCtx,
		cancel: cancel,
	}

	err := wp.Enqueue(taskRunning)
	assert.NoError(t, err)

	// Отмена выполнения через 100ms
	go func() {
		time.Sleep(100 * time.Millisecond)
		cancel()
	}()

	time.Sleep(300 * time.Millisecond)

	state, ok := wp.GetTaskState("task-running")
	assert.True(t, ok)
	assert.Equal(t, "failed", state)

	wp.Stop()
}

// SimpleTaskRunning эмулирует выполняющуюся и отменяемую задачу
type SimpleTaskRunning struct {
	id     string
	ctx    context.Context
	cancel context.CancelFunc
}

func (t *SimpleTaskRunning) Execute() error {
	select {
	case <-t.ctx.Done():
		return t.ctx.Err()
	case <-time.After(250 * time.Millisecond):
		return nil
	}
}

func (t *SimpleTaskRunning) ID() string {
	return t.id
}
