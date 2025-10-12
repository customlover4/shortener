package service

import (
	"errors"
	"fmt"
	"shortener/internal/entities/redirect"
	"shortener/internal/storage"
	"time"
)

type redirectsService struct {
	redirector

	redirects []redirect.Redirect
	i         int
}

func NewRedirects(r redirector) *redirectsService {
	return &redirectsService{
		redirector: r,
		redirects:  make([]redirect.Redirect, RedirectsBatchSize),
		i:          0,
	}
}

func (s *Service) CreateRedirect(r redirect.Redirect) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.rs.i != RedirectsBatchSize-1 {
		s.rs.redirects[s.rs.i] = r
		s.rs.i++
		return
	}

	s.rs.redirects[s.rs.i] = r
	tmp := make([]redirect.Redirect, len(s.rs.redirects))
	copy(tmp, s.rs.redirects)
	go s.rs.CreateRedirects(s.rs.redirects)
	s.rs.i = 0
}

func (s *Service) Redirects(alias string) ([]redirect.Redirect, error) {
	const op = "internal.service.redirects.Get"

	redirects, err := s.rs.redirector.Redirects(alias)
	if errors.Is(err, storage.ErrNotFound) {
		return nil, ErrNotFound
	} else if err != nil {
		return nil, fmt.Errorf(
			"%s: %w(%w)", op, ErrStorageInternal, err,
		)
	}

	return redirects, nil
}

func (s *Service) AgrigatedRedirects(opts redirect.AgrigateOpts) (redirect.Agrigated, error) {
	const op = "internal.service.redirects.AgrigatedGet"

	if opts.Alias == "" {
		return redirect.Agrigated{}, fmt.Errorf(
			"%w: %s", ErrNotValidData, "empty alias",
		)
	}
	if opts.StartDate != "" {
		_, err := time.Parse(time.DateTime, opts.StartDate)
		if err != nil {
			return redirect.Agrigated{}, fmt.Errorf(
				"%w: %s", ErrNotValidData,
				": start date, format 2006-01-02 15:04:05",
			)
		}
	}
	if opts.EndDate != "" {
		_, err := time.Parse(time.DateTime, opts.EndDate)
		if err != nil {
			return redirect.Agrigated{}, fmt.Errorf(
				"%w: %s", ErrNotValidData,
				": end date, format 2006-01-02 15:04:05",
			)
		}
	}

	res, err := s.rs.redirector.AgrigatedRedirects(opts)
	if errors.Is(err, storage.ErrNotFound) && res.Total == 0 {
		return redirect.Agrigated{}, ErrNotFound
	} else if err != nil {
		return redirect.Agrigated{}, fmt.Errorf(
			"%s: %w(%w)", op, ErrStorageInternal, err,
		)
	}

	return res, nil
}
