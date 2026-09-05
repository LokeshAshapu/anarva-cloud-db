package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/anarva-cloud/anarva-cloud-db/internal/worker"
)

func main() {
	log.Println("[ANARVA Compute Worker Data Plane] Starting worker service...")

	cfg, err := worker.LoadWorkerConfigFromEnv()
	if err != nil {
		log.Fatalf("FATAL: Failed to load worker configuration: %v", err)
	}

	store, err := worker.NewFileMetadataStore(cfg.DBPath)
	if err != nil {
		log.Fatalf("FATAL: Failed to initialize worker metadata store: %v", err)
	}

	runtime := worker.NewDockerContainerRuntime()
	if !runtime.HasRuntime() {
		log.Println("[ANARVA Compute Worker Warning] Host Docker runtime unavailable. Operating in simulation mode.")
	} else {
		log.Println("[ANARVA Compute Worker] Real Docker container runtime connected.")
	}

	srv, err := worker.NewWorkerServer(cfg, store, runtime)
	if err != nil {
		log.Fatalf("FATAL: Failed to initialize WorkerServer: %v", err)
	}

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		log.Printf("[ANARVA Compute Worker] Listening on %s (mTLS Client Auth: %v)...", cfg.ListenAddr, cfg.RequireClientAuth)
		if err := srv.Start(context.Background()); err != nil {
			log.Printf("[ANARVA Compute Worker] Server stopped: %v", err)
		}
	}()

	<-shutdown
	log.Println("[ANARVA Compute Worker] Shutting down worker service...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Stop(ctx); err != nil {
		log.Printf("ERROR: Shutdown failed: %v", err)
	} else {
		log.Println("[ANARVA Compute Worker] Graceful shutdown completed cleanly.")
	}
	fmt.Println("Goodbye!")
}
