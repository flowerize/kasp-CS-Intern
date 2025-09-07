package api

import (
	"encoding/json"
	"kasp-cs-tsvetkov-denis/internal/queue"
	"kasp-cs-tsvetkov-denis/internal/worker"
	"net/http"
)

type API interface {
	HealthHandler(w http.ResponseWriter, r *http.Request)
	EnqueueHandler(w http.ResponseWriter, r *http.Request)
}

type Server struct {
	taskQueue  queue.TaskQueueInterface
	workerPool worker.WorkerPoolInterface
}

func NewServer(taskQueue queue.TaskQueueInterface, workerPool worker.WorkerPoolInterface) API {
	return &Server{
		workerPool: workerPool,
		taskQueue:  taskQueue,
	}
}

func (s *Server) HealthHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok"))
}

func (s *Server) EnqueueHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	// Создаем структуру для парсинга JSON
	var request struct {
		ID         string `json:"id"`
		Payload    string `json:"payload"`
		MaxRetries int    `json:"max_retries"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid JSON: "+err.Error(), http.StatusBadRequest)
		return
	}

	task := worker.NewSimpleTask(request.ID, request.Payload, request.MaxRetries)

	err := s.workerPool.Enqueue(task)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusAccepted)
	w.Write([]byte("Task enqueued successfully"))
}
