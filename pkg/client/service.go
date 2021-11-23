package client

import "context"

// Service is a service interface.
type Service interface {
	// Add adds client.
	Add(ctx context.Context, client Entry) error
	// Send sends email to every Entry with given mailingID.
	// After that, Clients with that mailingID will be removed from db.
	Send(ctx context.Context, mailingID int) error
	// Delete deletes Entry with given params.
	Delete(ctx context.Context, id int) error
	// Get gets single Entry base on id.
	Get(ctx context.Context, id int) (*Entry, error)
	// List lists Clients with pagination.
	List(ctx context.Context, cursor Cursor) ([]Entry, error)
}

// service implements Service interface.
type service struct {
	repository Repository
}

// NewService returns new Service.
func NewService(repository Repository) Service {
	return service{repository: repository}
}

func (s service) Add(ctx context.Context, client Entry) error {
	return s.repository.Insert(ctx, client)
}

func (s service) Send(ctx context.Context, mailingID int) error {
	clients, err := s.repository.GetFilter(ctx, &getParams{mailingID: &mailingID})
	if err != nil {
		return err
	}

	// Do the mailing things. Loop through all of clients and send emails using some kind of interface dependency.

	ids := make([]int, 0)
	for _, c := range clients {
		ids = append(ids, c.ID)
	}

	return s.repository.BatchDelete(ctx, ids)
}

func (s service) Delete(ctx context.Context, id int) error {
	return s.repository.Delete(ctx, id)
}

func (s service) Get(ctx context.Context, id int) (*Entry, error) {
	return s.repository.Get(ctx, id)
}

func (s service) List(ctx context.Context, cursor Cursor) ([]Entry, error) {
	return s.repository.GetFilter(ctx, &getParams{
		offset: cursor.AfterID,
		limit:  &cursor.Limit,
	})
}
