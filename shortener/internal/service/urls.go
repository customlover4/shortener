package service

import (
	"errors"
	"fmt"
	parser "net/url"
	"shortener/internal/entities/url"
	"shortener/internal/storage"
)

const (
	AliasLen = 6
)

func (s *Service) CreateURL(u url.URL) (string, error) {
	const op = "internal.service.url.Create"

	var err error

	if u.Original == "" {
		return "", fmt.Errorf("%w: %s", ErrNotValidData, "empty original link")
	}
	original, err := parser.Parse(u.Original)
	if err != nil || original.Host == "" || original.Scheme == "" {
		return "", fmt.Errorf("%w: %s", ErrNotValidData, "url is not valid (scheme://host/path)")
	}

	genAlias := false
	if u.Alias == "" {
		u.Alias = generateAlias(AliasLen)
		genAlias = true
	}

	var alias string
	for i := 0; i < 10; i++ {
		alias, err = s.urler.CreateURL(u)
		if errors.Is(err, storage.ErrNotUnique) && genAlias {
			u.Alias = generateAlias(AliasLen + i)
			continue
		} else if errors.Is(err, storage.ErrNotUnique) {
			return alias, ErrNotUnique
		} else if err != nil {
			return alias, fmt.Errorf("%s: %w(%w)", op, ErrStorageInternal, err)
		}

		break
	}

	return alias, nil
}

func (s *Service) URL(alias string) (string, error) {
	const op = "internal.service.url.Get"

	if alias == "" {
		return "", fmt.Errorf(
			"%w: %s", ErrNotValidData, "empty alias",
		)
	}

	u, err := s.urler.URL(alias)
	if errors.Is(err, storage.ErrNotFound) {
		return u, ErrNotFound
	} else if err != nil {
		return u, fmt.Errorf("%s: %w(%w)", op, ErrStorageInternal, err)
	}

	return u, nil
}
