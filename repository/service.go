package repository

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
)

type (
	// Service is an interface which should implement the
	// actual business logic for repositories.
	Service interface {
		Get(ctx context.Context, url string) (Repository, error)
		Homepage(ctx context.Context) (Homepage, error)
	}
	// Storage is an interface which implementation should actually
	// store and retrieve repositories.
	Storage interface {
		Get(ctx context.Context, url string) (Repository, error)
		GetPopular(ctx context.Context, limit int) ([]string, error)
		GetLatest(ctx context.Context, limit int) ([]string, error)
		GetRandom(ctx context.Context, limit int) ([]string, error)
		Exists(ctx context.Context, url string) (bool, error)
		Create(ctx context.Context, repo Repository) error
	}
)

// ErrNotFound is returned when a Repository was not found
var ErrNotFound = errors.New("repository not found")

type service struct {
	github       *GitHub
	godoc        *GoDoc
	repositories Storage
}

// NewService creates a new Service implementation which works with a Storage.
func NewService(repositories Storage, gh *GitHub, gd *GoDoc) Service {
	return &service{
		github:       gh,
		godoc:        gd,
		repositories: repositories,
	}
}

func (s *service) Get(ctx context.Context, url string) (Repository, error) {
	exists, err := s.repositories.Exists(ctx, url)
	if err != nil {
		return Repository{}, err
	}

	if !exists {
		godocInfo, err := s.godoc.Get(ctx, url)
		if err != nil {
			return Repository{}, err
		}

		repo, err := s.github.Get(ctx, url)
		if err != nil {
			return repo, err
		}

		if godocInfo.Imports > 0 {
			repo.Statistics = append(repo.Statistics, Statistic{
				Name:  "Imports",
				Value: godocInfo.Imports,
				URL:   fmt.Sprintf("https://godoc.org/%s?imports", repo.URL),
			})
		}
		if godocInfo.Importers > 0 {
			repo.Statistics = append(repo.Statistics, Statistic{
				Name:  "Importers",
				Value: godocInfo.Importers,
				URL:   fmt.Sprintf("https://godoc.org/%s?importers", repo.URL),
			})
		}

		if err := s.repositories.Create(ctx, repo); err != nil {
			return repo, err
		}
	}

	repo, err := s.repositories.Get(ctx, url)
	if err != nil && err != ErrNotFound {
		return repo, err
	}

	return repo, err
}

// Homepage contains urls of repositories with different categories
type Homepage struct {
	Popular []string
	Latest  []string
	Random  []string
}

func (s *service) Homepage(ctx context.Context) (Homepage, error) {
	limit := 15

	// TODO: Use a sync.WaitGroup to run this concurrently

	h := Homepage{}

	popular, err := s.repositories.GetPopular(ctx, limit)
	if err != nil {
		return h, err
	}

	latest, err := s.repositories.GetLatest(ctx, limit)
	if err != nil {
		return h, err
	}

	random, err := s.repositories.GetRandom(ctx, limit)
	if err != nil {
		return h, err
	}

	return Homepage{
		Popular: popular,
		Latest:  latest,
		Random:  random,
	}, nil
}
