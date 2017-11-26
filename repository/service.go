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
	}
	// Storage is an interface which implementation should actually
	// store and retrieve repositories.
	Storage interface {
		Get(ctx context.Context, url string) (Repository, error)
		Exists(ctx context.Context, url string) (bool, error)
		Create(ctx context.Context, repo Repository) error
	}
)

var NotFoundErr = errors.New("repository not found")

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
			repo.Stats = append(repo.Stats, Stat{
				Name:  "Imports",
				Value: godocInfo.Imports,
				URL:   fmt.Sprintf("https://godoc.org/%s?imports", repo.URL),
			})
		}
		if godocInfo.Importers > 0 {
			repo.Stats = append(repo.Stats, Stat{
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
	if err != nil && err != NotFoundErr {
		return repo, err
	}

	return repo, err
}
