package worker

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"sync"
	"time"
)

type Task interface {
	Execute() error
	ID() string
}

type SimpleTask struct {
	TaskID     string `json:"id"`
	Payload    string `json:"payload"`
	MaxRetries int    `json:"max_retries"`
}

func (t *SimpleTask) ID() string {
	return t.TaskID
}

func NewSimpleTask(id, payload string, maxRetries int) *SimpleTask {
	return &SimpleTask{
		TaskID:     id,
		Payload:    payload,
		MaxRetries: maxRetries,
	}
}

type WorkerPoolInterface interface {
	Enqueue(task Task) error
	Start()
	Stop()
	GetTaskState(taskID string) (string, bool)
}

type WorkerPool struct {
	workers   int
	taskQueue chan Task
	wg        sync.WaitGroup
	ctx       context.Context
	cancel    context.CancelFunc
	states    sync.Map // map[string]string - состояние задач
}

func NewWorkerPool(workers int, queueSize int) *WorkerPool {
	ctx, cancel := context.WithCancel(context.Background())
	return &WorkerPool{
		workers:   workers,
		taskQueue: make(chan Task, queueSize),
		ctx:       ctx,
		cancel:    cancel,
	}
}

// Start запускает воркеры
func (wp *WorkerPool) Start() {
	for i := 0; i < wp.workers; i++ {
		wp.wg.Add(1)
		go func() {
			wp.worker()
		}()
	}
}

// Stop останавливает пул и ждёт завершения воркеров
func (wp *WorkerPool) Stop() {
	wp.cancel()
	close(wp.taskQueue)
	wp.wg.Wait()
}

// Enqueue добавляет задачу в очередь
func (wp *WorkerPool) Enqueue(task Task) error {
	select {
	case <-wp.ctx.Done():
		return context.Canceled
	case wp.taskQueue <- task:
		wp.states.Store(task.ID(), "queued")
		return nil
	}
}

// GetTaskState возвращает состояние задачи
func (wp *WorkerPool) GetTaskState(taskID string) (string, bool) {
	v, ok := wp.states.Load(taskID)
	if !ok {
		return "", false
	}
	return v.(string), true
}

func (wp *WorkerPool) worker() {
	defer wp.wg.Done()
	for {
		select {
		case <-wp.ctx.Done():
			return
		case task, ok := <-wp.taskQueue:
			if !ok {
				return
			}
			wp.states.Store(task.ID(), "running")
			err := task.Execute()
			if err != nil {
				wp.states.Store(task.ID(), "failed")
			} else {
				wp.states.Store(task.ID(), "done")
			}
		}
	}
}

// Execute Симуляция выполнения задачи
func (t *SimpleTask) Execute() error {
	var attempt int
	baseDelay := 100 * time.Millisecond
	maxDelay := 2 * time.Second

	for {
		attempt++
		delay := time.Duration(100+rand.Intn(400)) * time.Millisecond
		time.Sleep(delay)

		// Симуляция падения задачи с вероятностью 20%
		if rand.Float64() < 0.2 {
			err := errors.New("task failed (simulated)")
			fmt.Printf("Task %s attempt %d failed\n", t.TaskID, attempt)
			if attempt > t.MaxRetries {
				fmt.Printf("Task %s reached max retries\n", t.TaskID)
				return err
			}
			backoff := baseDelay << (attempt - 1)
			if backoff > maxDelay {
				backoff = maxDelay
			}
			jitter := time.Duration(rand.Int63n(int64(backoff) / 2))
			backoff = backoff - jitter
			time.Sleep(backoff)

			continue
		}

		fmt.Printf("Task %s executed successfully on attempt %d\n", t.TaskID, attempt)
		return nil
	}
}
