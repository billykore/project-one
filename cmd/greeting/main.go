package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/billykore/project-one/internal/app/greeting/adapters/handler"
	"github.com/billykore/project-one/internal/app/greeting/adapters/repository"
	"github.com/billykore/project-one/internal/app/greeting/core/service"
	"github.com/billykore/project-one/pkg/logger"
)

func main() {
	// Initialize logger
	log := logger.New()

	// Initialize driven adapters (infrastructure)
	repo := repository.NewMemoryGreetingRepository()

	// Initialize core business logic (domain + service)
	greetingService := service.NewGreetingService(log, repo)

	// Initialize driving adapters (entry points)
	greetingHandler := handler.NewGreetingHandler(log, greetingService)

	// Set up standard HTTP ServeMux
	mux := http.NewServeMux()
	mux.HandleFunc("/greeting", greetingHandler.GetGreeting)

	// Configure server
	server := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}

	// Start server in a goroutine
	go func() {
		log.Info("starting server", "addr", server.Addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error("server failed to start", "error", err)
			os.Exit(1)
		}
	}()

	// Graceful shutdown: wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Info("received shutdown signal, shutting down gracefully...")

	// Allow 10 seconds for active requests to complete
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Error("server forced to shutdown", "error", err)
		os.Exit(1)
	}

	log.Info("server exited gracefully")
}
