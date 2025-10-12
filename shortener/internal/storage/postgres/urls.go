package postgres

import (
	"context"
	"fmt"
	"shortener/internal/entities/url"
)

func (p *Postgres) CreateURL(u url.URL) (string, error) {
	p.semaphore <- struct{}{}
	defer func() { <-p.semaphore }()

	const op = "internal.storage.postgres.url.Create"

	q := fmt.Sprintf("insert into %s (alias, original) values ($1, $2);", URLTable)

	_, err := p.db.ExecContext(context.Background(), q, u.Alias, u.Original)
	if err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}

	return u.Alias, nil
}

func (p *Postgres) URL(alias string) (string, error) {
	p.semaphore <- struct{}{}
	defer func() { <-p.semaphore }()
	
	const op = "internal.storage.postgres.url.Get"

	var u string

	q := fmt.Sprintf("select original from %s where alias = $1;", URLTable)
	rows := p.db.Master.QueryRow(q, alias)
	if rows.Err() != nil {
		return u, fmt.Errorf("%s: %w", op, rows.Err())
	}
	err := rows.Scan(&u)
	if err != nil {
		return u, fmt.Errorf("%s: %w", op, err)
	}

	return u, nil
}
