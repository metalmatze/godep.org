package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/pkg/errors"
)

type postgres struct {
	db *sql.DB
}

// NewPostgresStorage returns a Storage implementation using Postgres.
func NewPostgresStorage(db *sql.DB) Storage {
	return &postgres{db: db}
}

func (p *postgres) Get(ctx context.Context, url string) (Repository, error) {
	var r Repository
	var id string
	{
		q := "SELECT id, url, description, updated FROM repositories " +
			"WHERE url = $1 LIMIT 1;"
		row := p.db.QueryRowContext(ctx, q, url)

		err := row.Scan(&id, &r.URL, &r.Description, &r.Updated)
		if err != nil && err.Error() == "sql: no rows in result set" {
			return r, NotFoundErr
		}
	}

	// Fetch all repository statistics
	{
		q := "SELECT name, value, url FROM statistics " +
			"WHERE repository_id = $1 ORDER BY name ASC;"
		rows, err := p.db.QueryContext(ctx, q, id)
		if err != nil {
			return r, errors.Wrap(err, "failed to fetch repository statistics")
		}
		defer rows.Close()

		for rows.Next() {
			s := Stat{}
			if err := rows.Scan(&s.Name, &s.Value, &s.URL); err != nil {
				return r, errors.Wrap(err, "failed to scan repository stat")
			}
			r.Stats = append(r.Stats, s)
		}
		if err := rows.Err(); err != nil {
			return r, errors.Wrap(err, "failed to retrieve repository statistics")
		}
	}
	// Fetch all repository versions
	{
		q := "SELECT name, published FROM versions WHERE repository_id = $1 ORDER BY sort_order DESC LIMIT 25"
		rows, err := p.db.QueryContext(ctx, q, id)
		if err != nil {
			return r, errors.Wrap(err, "failed to fetch repository versions")
		}
		defer rows.Close()

		for rows.Next() {
			var published *time.Time
			v := Version{}
			if err := rows.Scan(&v.Name, &published); err != nil {
				return r, errors.Wrap(err, "failed to scan repository version")
			}
			if published != nil {
				v.Published = *published
			}

			fmt.Printf("%+v\n", v)
			r.Versions = append(r.Versions, v)
		}
		if err := rows.Err(); err != nil {
			return r, errors.Wrap(err, "failed to retrieve repository versions")
		}

		if len(r.Versions) > 0 {
			r.CurrentVersion = r.Versions[0]
		}
	}

	return r, nil
}

func (p *postgres) Exists(ctx context.Context, url string) (bool, error) {
	q := `SELECT url FROM repositories WHERE url = $1 LIMIT 1`
	row := p.db.QueryRowContext(ctx, q, url)

	var u string
	err := row.Scan(&u)
	if err != nil && err.Error() != "sql: no rows in result set" {
		return false, err
	}

	return url == u, nil
}

func (p *postgres) Create(ctx context.Context, repo Repository) error {
	tx, err := p.db.BeginTx(ctx, nil)
	if err != nil {
		return errors.Wrap(err, "failed to create transaction")
	}

	var id string
	{
		q := `INSERT INTO repositories (url, description, updated) VALUES ($1, $2, $3) RETURNING id`
		row := tx.QueryRowContext(ctx, q, repo.URL, repo.Description, repo.Updated)

		if err := row.Scan(&id); err != nil {
			tx.Rollback()
			return errors.Wrap(err, "failed to scan repository id")
		}
	}

	// statistics
	{
		q := `INSERT INTO statistics (repository_id, name, value, url) VALUES ($1, $2, $3, $4)`
		stmt, err := tx.PrepareContext(ctx, q)
		if err != nil {
			return errors.Wrap(err, "failed to prepare the inserting statistics query")
		}
		defer stmt.Close()

		for _, stat := range repo.Stats {
			if _, err := stmt.ExecContext(ctx, id, stat.Name, stat.Value, stat.URL); err != nil {
				tx.Rollback()
				return errors.Wrap(err, "failed to insert repository stat")
			}
		}
	}

	// versions
	{
		q := `INSERT INTO versions (repository_id, name, sort_order, published) VALUES ($1, $2, $3, $4)`
		stmt, err := tx.PrepareContext(ctx, q)
		if err != nil {
			return errors.Wrap(err, "failed to prepare the inserting versions query")
		}
		defer stmt.Close()

		for i, v := range repo.Versions {
			var published *time.Time
			if !v.Published.IsZero() {
				published = &v.Published
			}

			if _, err := stmt.ExecContext(ctx, id, v.Name, i, published); err != nil {
				tx.Rollback()
				return errors.Wrap(err, "failed to insert repository versions")
			}
		}
	}

	if err := tx.Commit(); err != nil {
		return errors.Wrap(err, "failed to commit transaction")
	}

	return nil
}
