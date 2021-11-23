package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
	"vodeno/pkg/client"
	"vodeno/pkg/config"
	db2 "vodeno/pkg/db"
	"vodeno/pkg/middleware"

	"github.com/go-chi/chi/v5"
	_ "github.com/lib/pq"
	"github.com/sirupsen/logrus"
)

func main() {
	ctx := context.Background()
	logger := logrus.New()

	r := chi.NewRouter()
	r.Use(
		middleware.LoggerMiddleware(logger),
		middleware.AuthenticationMiddleware,
	)
	cfg, err := config.Load()
	if err != nil {
		logger.Panic(err)
	}

	db, err := db2.OpenDB(cfg.DB)
	if err != nil {
		logger.Panic(err)
	}

	repo := client.NewRepo(db)
	service := client.NewService(repo)
	handler := client.NewHandler(logger, service)

	watcher := client.NewWatcher(logger, repo, cfg.Watcher.TickPeriod)
	watcher.Start(ctx)

	handler.AddRoutes(r)

	pid := os.Getpid()
	srvAddr := fmt.Sprintf(":%d", cfg.Port)

	srv := newServer(srvAddr, r)
	go func() {
		logger.WithFields(logrus.Fields{
			"PID":  pid,
			"addr": srvAddr,
		}).Info("starting main API srv")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.WithError(err).Error("starting main API srv failed")
		}
	}()

	// Graceful shutdown.
	term := make(chan os.Signal, 1)
	signal.Notify(term, os.Interrupt, syscall.SIGTERM)

	<-term
	watcher.Stop()
	if err := srv.Shutdown(ctx); err != nil {
		logger.WithError(err).Error("shutting down API failed")
	}
	logrus.Info("exiting")
}

// newServer creates a new http.Server object.
func newServer(serverAddr string, handler http.Handler) *http.Server {
	srv := &http.Server{
		Addr:         serverAddr,
		ReadTimeout:  600 * time.Second,
		WriteTimeout: 900 * time.Second,
		IdleTimeout:  600 * time.Second,
		Handler:      handler,
	}
	return srv
}
