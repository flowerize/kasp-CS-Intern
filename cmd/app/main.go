package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"kasp-cs-tsvetkov-denis/internal/api"
	"kasp-cs-tsvetkov-denis/internal/config"
	"kasp-cs-tsvetkov-denis/internal/queue"
	"kasp-cs-tsvetkov-denis/internal/worker"
)

func main() {
	cfg := config.NewConfig()

	// Создание очереди и пула воркеров
	taskQueue := queue.New[worker.Task](cfg.QueueSize)
	workerPool := worker.NewWorkerPool(cfg.Workers, cfg.QueueSize)
	workerPool.Start()

	// Создаем сервер и HTTP обработчики
	srv := api.NewServer(taskQueue, workerPool)

	mux := http.NewServeMux()
	mux.HandleFunc("/enqueue", srv.EnqueueHandler)
	mux.HandleFunc("/healthz", srv.HealthHandler)

	httpServer := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	// Канал для сигнала завершения
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		log.Println("Server started on :8080")
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("HTTP server ListenAndServe error: %v", err)
		}
	}()

	<-stop
	log.Println("Shutdown signal received")

	taskQueue.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := httpServer.Shutdown(ctx); err != nil {
		log.Fatalf("HTTP server Shutdown error: %v", err)
	}

	workerPool.Stop()

	log.Println("Server gracefully stopped")
}
