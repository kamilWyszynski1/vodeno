package db

import (
	"errors"
	"os"
	"os/signal"
	"syscall"
	"time"
	"vodeno/pkg/config"

	"github.com/jmoiron/sqlx"
)

var (
	// ErrIsReadyDBTimeout is returned when connection to the database couldn't be established in specified period of time.
	ErrIsReadyDBTimeout = errors.New("DB ready check timeout")
	// ErrIsReadyDBTerminated is returned when database connection check is terminated by SIGTERM signal.
	ErrIsReadyDBTerminated = errors.New("DB ready check received SIGTERM")
)

// OpenDBWithTimeout waits for the specified time until database connection is ready.
func OpenDBWithTimeout(cfg config.DBConfig, waitForDBSecs int) (*sqlx.DB, error) {
	// Connect immediately when waitForDBSecs is not set.
	if waitForDBSecs < 1 {
		return OpenDB(cfg)
	}

	// Graceful shutdown
	term := make(chan os.Signal, 1)
	signal.Notify(term, os.Interrupt, syscall.SIGTERM)
	defer signal.Stop(term)

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	t := time.Now()

	for {
		// Try to establish connection.
		db, err := OpenDB(cfg)
		if err == nil {
			return db, nil
		}
		if time.Since(t) > time.Duration(waitForDBSecs)*time.Second {
			return nil, ErrIsReadyDBTimeout
		}
		// Wait some time.
		select {
		case <-term:
			return nil, ErrIsReadyDBTerminated
		case <-ticker.C:
		}
	}
}

// OpenDB opens postgres database.
func OpenDB(cfg config.DBConfig) (*sqlx.DB, error) {
	db, err := sqlx.Open("postgres", cfg.ConnectionString())
	if err != nil {
		return nil, err
	}

	db.SetMaxOpenConns(cfg.MaxOpenConns)
	db.SetMaxIdleConns(cfg.MaxIdleConns)
	db.SetConnMaxLifetime(time.Second * time.Duration(cfg.ConnMaxLifetimeSecs))

	err = db.Ping()
	return db, err
}
