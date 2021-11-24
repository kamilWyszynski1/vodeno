package client

import (
	"context"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

var ttl = time.Minute * 5 // time for client entry to exist.

// Watcher is responsible for watching over old entries and deleting them.
type Watcher struct {
	tickPeriod time.Duration
	log        logrus.FieldLogger
	wg         sync.WaitGroup
	repo       Repository
	close      chan struct{} // channel is used for graceful shutdown
}

// NewWatcher create new instance of Watcher.
func NewWatcher(logger *logrus.Logger, repo Repository, tp time.Duration) *Watcher {
	return &Watcher{
		log:        logger.WithField("place", "watcher"),
		repo:       repo,
		close:      make(chan struct{}),
		tickPeriod: tp,
	}
}

// Start starts clearing goroutine.
func (w *Watcher) Start(ctx context.Context) {
	w.wg.Add(1)
	w.log.WithField("tick", w.tickPeriod.String()).Info("starting")
	go func() {
		defer w.wg.Done()
		ticker := time.NewTicker(w.tickPeriod)
		for {
			select {
			case <-ticker.C:
				w.log.Info("clearing")
				if err := w.clear(ctx); err != nil {
					w.log.WithError(err).Error("failed to clear old client entries.")
				}
			case <-w.close:
				w.log.Info("closing")
				return
			}
		}
	}()
}

// clear queries repository for entries to be deleted and deletes them.
func (w *Watcher) clear(ctx context.Context) error {
	t := time.Now().Add(-ttl)
	clients, err := w.repo.GetFilter(ctx, &getParams{insertTimeLt: &t})
	if err != nil {
		return err
	}

	ids := make([]int, 0, len(clients))
	for _, client := range clients {
		ids = append(ids, client.ID)
	}
	return w.repo.BatchDelete(ctx, ids)
}

// Stop stops watcher and waits for goroutine to shutdown.
func (w *Watcher) Stop() {
	w.close <- struct{}{}
	w.wg.Wait()
}
