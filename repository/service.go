package repository

import "context"

type (
	// Service is an interface which should implement the
	// actual business logic for repositories.
	Service interface {
		Get(ctx context.Context, url string) (Repository, error)
	}
	// Storage is an interface which implementation should actually
	// store and retrieve repositories.
	Storage interface {
		Get(ctx context.Context, url string) (Repository, error)
	}
)

type service struct {
	repositories Storage
}

// NewService creates a new Service implementation which works with a Storage.
func NewService(repositories Storage) (Service, error) {
	return &service{
		repositories: repositories,
	}, nil
}

func (s *service) Get(ctx context.Context, url string) (Repository, error) {
	return s.repositories.Get(ctx, url)
}
