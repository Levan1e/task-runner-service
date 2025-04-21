package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"task-runner-service/internal/api"
	"task-runner-service/internal/config"
	"task-runner-service/internal/service"
	"task-runner-service/internal/storage/redis"
	"task-runner-service/pkg/logger"

	v1 "task-runner-service/internal/api/v1"

	"github.com/RichardKnop/machinery/v1"
	machineryConfig "github.com/RichardKnop/machinery/v1/config"
)

func main() {
	absPath, err := filepath.Abs("./internal/config/config.yaml")
	if err != nil {
		log.Fatalf("Failed to resolve config file path: %v", err)
	}

	cfg, err := config.ParseConfig(absPath)
	if err != nil {
		logger.Errorf("Error parsing configuration: %v", err)
		log.Fatal("Exiting due to configuration error")
	}
	logger.Infof("Configuration loaded: %+v", cfg)

	redisStorage, err := redis.NewStorage(*cfg.Redis)
	if err != nil {
		logger.Errorf("Error initializing Redis storage: %v", err)
		log.Fatal("Exiting due to Redis initialization error")
	}
	logger.Info("Redis storage initialized successfully")

	machineryCfg := &machineryConfig.Config{
		Broker:        cfg.Broker.Broker,
		DefaultQueue:  cfg.Broker.DefaultQueue,
		ResultBackend: cfg.Broker.ResultBackend,
	}
	machineryServer, err := machinery.NewServer(machineryCfg)
	if err != nil {
		logger.Errorf("Error creating Machinery server: %v", err)
		log.Fatal("Exiting due to Machinery server error")
	}
	logger.Info("Machinery server created")

	go func() {
		if err := runWorkers(machineryServer); err != nil {
			logger.Errorf("Error starting workers: %v", err)
			log.Fatal("Exiting due to worker startup error")
		}
	}()

	runnerService := service.NewRunnerService(machineryServer, redisStorage)
	v1Handler := v1.NewHandler(runnerService)

	httpConfig := &api.HTTPConfig{
		Host:         cfg.Server.Host,
		Port:         cfg.Server.Port,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
	}
	server := api.NewServer(httpConfig, v1Handler)

	go func() {
		if err := server.Run(); err != nil && err != http.ErrServerClosed {
			logger.Errorf("Error starting HTTP server: %v", err)
			log.Fatal("Exiting due to HTTP server error")
		}
	}()
	logger.Infof("HTTP server listening on port %s", httpConfig.Port)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down application...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Stop(ctx); err != nil {
		logger.Errorf("Error stopping HTTP server: %v", err)
		log.Fatal("Exiting due to HTTP shutdown error")
	}
}

func runWorkers(server *machinery.Server) error {
	worker := server.NewWorker("task_worker", 10)
	if err := worker.Launch(); err != nil {
		return fmt.Errorf("failed to launch worker: %w", err)
	}
	return nil
}
