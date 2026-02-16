package main

import (
	"context"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/example/fanji-backend/internal/app"
	"github.com/example/fanji-backend/internal/config"
	"github.com/example/fanji-backend/internal/messaging"
	"github.com/example/fanji-backend/internal/storage"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	redisStore := storage.NewRedisStore(cfg.RedisAddr, cfg.RedisPass)
	mysqlStore, err := storage.NewMySQLStore(cfg.MySQLDSN)
	if err != nil {
		log.Fatalf("init mysql: %v", err)
	}
	publisher, err := messaging.NewKafkaPublisher(cfg.KafkaBrokers, cfg.KafkaTopic)
	if err != nil {
		log.Fatalf("init kafka: %v", err)
	}

	server := app.NewServer(redisStore, mysqlStore, publisher)
	httpServer := &http.Server{
		Addr:    cfg.HTTPAddr,
		Handler: server.Handler(),
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
		defer cancel()
		_ = httpServer.Shutdown(shutdownCtx)
		_ = server.Close(shutdownCtx)
	}()

	log.Printf("server listening on %s", cfg.HTTPAddr)
	if err = httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("listen: %v", err)
	}
}
