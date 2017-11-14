package repository

import (
	"context"
	"database/sql"
)

type storage struct {
	db *sql.DB
}

// NewStorage returns a Storage implementation using Postgres.
func NewStorage(db *sql.DB) (Storage, error) {
	return &storage{db: db}, nil
}

func (s *storage) Get(ctx context.Context, url string) (Repository, error) {
	q := `
SELECT
  url,
  description,
  updated
FROM repositories
WHERE url = $1;
	`
	row := s.db.QueryRowContext(ctx, q, url)

	var r Repository
	err := row.Scan(&r.URL, &r.Description, &r.Updated)
	return r, err
}
