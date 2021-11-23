package client

import (
	"context"
	"time"
)

// getParams is a container for GetFilter method filtering.
// If needed we can add another filters.
type getParams struct {
	mailingID    *int
	insertTimeLt *time.Time
	offset       *int
	limit        *int
}

// Repository is a repository interface.
type Repository interface {
	// Insert inserts Entry to storage.
	Insert(ctx context.Context, c Entry) error
	// Delete deletes Entry from storage.
	Delete(ctx context.Context, id int) error
	// BatchDelete delete multiple Clients at once.
	BatchDelete(ctx context.Context, ids []int) error
	// GetFilter gets Entries from storage.
	// If params are nil it gets all Entries.
	GetFilter(ctx context.Context, params *getParams) ([]Entry, error)
	// Get queries single Entry.
	Get(ctx context.Context, id int) (*Entry, error)
}
