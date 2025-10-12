package service

import (
	"errors"
	"shortener/internal/entities/redirect"
	"shortener/internal/entities/url"
	"sync"
)

const (
	RedirectsBatchSize = 250
)

var (
	ErrNotUnique       = errors.New("not unique value")
	ErrStorageInternal = errors.New("internal storage error")
	ErrNotFound        = errors.New("not found url")
	ErrNotValidData    = errors.New("not valid data")
)

type urler interface {
	CreateURL(u url.URL) (string, error)
	URL(alias string) (string, error)
}

type redirector interface {
	CreateRedirects(redirects []redirect.Redirect)
	Redirects(alias string) ([]redirect.Redirect, error)
	AgrigatedRedirects(opts redirect.AgrigateOpts) (redirect.Agrigated, error)
}

type Service struct {
	urler
	rs *redirectsService

	mu *sync.Mutex
}

func New(u urler, r redirector) *Service {
	return &Service{
		urler: u,
		rs:    NewRedirects(r),
		mu:    new(sync.Mutex),
	}
}

func (s *Service) Shutdown() {
	if s.rs.i != 0 {
		s.rs.CreateRedirects(s.rs.redirects[:s.rs.i])
	}
}
