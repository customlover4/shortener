package storage

import (
	"database/sql"
	"errors"
	"fmt"
	"shortener/internal/entities/redirect"
	"shortener/internal/entities/url"
	"shortener/internal/storage/postgres"

	"github.com/wb-go/wbf/zlog"
)

var (
	ErrNotUnique = errors.New("not unique value for field")
	ErrNotFound  = errors.New("not found row")
)

type db interface {
	CreateURL(u url.URL) (string, error)
	URL(alias string) (string, error)
	CreateRedirects(redirects []redirect.Redirect)
	Redirects(alias string) ([]redirect.Redirect, error)
	AgrigatedRedirects(opts redirect.AgrigateOpts) (redirect.Agrigated, error)

	HandleError(err error) error
	Shutdown()
}

type cache interface {
	AddURL(alias, original string) error
	URL(alias string) (string, error)
	Shutdown()
}

type Storage struct {
	db db
	c  cache
}

func New(db db, c cache) *Storage {
	return &Storage{
		db: db,
		c:  c,
	}
}

func (s *Storage) CreateURL(u url.URL) (string, error) {
	const op = "internal.storage.createURL"
	alias, err := s.db.CreateURL(u)
	DBErr := s.db.HandleError(err)
	if errors.Is(DBErr, postgres.ErrNotUnique) {
		return "", ErrNotUnique
	} else if DBErr != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	} else if err != nil {
		return "", fmt.Errorf("%s: %w", op, err)
	}

	return alias, nil
}

func (s *Storage) URL(alias string) (string, error) {
	const op = "internal.storage.GetURL"

	original, err := s.c.URL(alias)
	if err != nil {
		return original, fmt.Errorf("%s: %w", op, err)
	}
	if original != "" {
		return original, nil
	}

	original, err = s.db.URL(alias)
	if errors.Is(err, sql.ErrNoRows) {
		return original, ErrNotFound
	} else if err != nil {
		return original, fmt.Errorf("%s: %w", op, err)
	}

	err = s.c.AddURL(alias, original)
	if err != nil {
		zlog.Logger.Error().Err(err).Fields(map[string]any{"op": op}).Send()
	}

	return original, nil
}

func (s *Storage) CreateRedirects(redirects []redirect.Redirect) {
	s.db.CreateRedirects(redirects)
}

func (s *Storage) Redirects(alias string) ([]redirect.Redirect, error) {
	const op = "internal.storage.GetRedirects"

	redirects, err := s.db.Redirects(alias)
	if errors.Is(err, sql.ErrNoRows) {
		return redirects, ErrNotFound
	} else if err != nil {
		return redirects, fmt.Errorf("%s: %w", op, err)
	}

	return redirects, nil
}

func (s *Storage) AgrigatedRedirects(opts redirect.AgrigateOpts) (redirect.Agrigated, error) {
	const op = "internal.storage.AgrigatedRedirects"

	res, err := s.db.AgrigatedRedirects(opts)
	if errors.Is(err, sql.ErrNoRows) || res.Total == 0 {
		return redirect.Agrigated{}, ErrNotFound
	} else if err != nil {
		return redirect.Agrigated{}, fmt.Errorf("%s: %w", op, err)
	}

	return res, nil
}

func (s *Storage) Shutdown() {
	s.c.Shutdown()
	s.db.Shutdown()
}
