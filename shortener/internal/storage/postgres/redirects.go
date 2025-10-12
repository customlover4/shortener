package postgres

import (
	"context"
	"fmt"
	"shortener/internal/entities/redirect"
	"strings"
	"sync"

	"github.com/wb-go/wbf/zlog"
)

func (p *Postgres) CreateRedirects(tmp []redirect.Redirect) {
	p.semaphore <- struct{}{}
	defer func() { <-p.semaphore }()

	const op = "internal.storage.postgres.redirect.CreateBatch"

	vals := make([]any, 0, len(tmp)*3)
	q := strings.Builder{}

	q.Grow(len(tmp)*12 + 52)
	q.WriteString(
		fmt.Sprintf("insert into %s (alias, dt, user_agent) values",
			RedirectsTable,
		),
	)
	counter := 1
	for _, r := range tmp {
		q.WriteString(
			fmt.Sprintf(" ($%d, $%d, $%d), ", counter, counter+1, counter+2),
		)
		counter += 3
		vals = append(vals, r.Alias, r.Date, r.UserAgent)
	}
	s := q.String()
	_, err := p.db.ExecContext(
		context.Background(), s[:len(s)-2], vals...,
	)
	if err != nil {
		zlog.Logger.Error().Err(err).Msg(op)
	}
}

func (p *Postgres) Redirects(alias string) ([]redirect.Redirect, error) {
	p.semaphore <- struct{}{}
	defer func() { <-p.semaphore }()

	const op = "internal.storage.postgres.redirect.Get"

	q := fmt.Sprintf("select * from %s where alias = $1;", RedirectsTable)
	rows, err := p.db.Master.Query(q, alias)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	defer func() {
		_ = rows.Close()
	}()

	var res []redirect.Redirect
	for rows.Next() {
		var tmp redirect.Redirect
		err := rows.Scan(&tmp.ID, &tmp.Alias, &tmp.Date, &tmp.UserAgent)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}
		res = append(res, tmp)
	}

	return res, nil
}

func (p *Postgres) generateAgrigatedReq(opts redirect.AgrigateOpts) (q string, args []any) {
	q = fmt.Sprintf("select * from %s where alias = $1", RedirectsTable)

	i := 2
	args = make([]any, 0, 4)
	args = append(args, opts.Alias)
	if opts.StartDate != "" {
		q += fmt.Sprintf(" and dt >= $%d", i)
		i++
		args = append(args, opts.StartDate)
	} else {
		q += " and dt >= '100-10-22 00:00:00'"
	}
	if opts.EndDate != "" {
		q += fmt.Sprintf(" and dt <= $%d", i)
		i++
		args = append(args, opts.EndDate)
	} else {
		q += " and dt <= '100000-10-22 00:00:00'"
	}
	if opts.FilterColumn == redirect.FilterUserAgent && opts.ValueForFilter != "" {
		q += fmt.Sprintf(" and STRPOS(%s, $%d) > 0", redirect.FilterUserAgent, i)
		i++
		args = append(args, opts.ValueForFilter)
	}
	if opts.Page == 0 {
		opts.Page++
	}
	offset := (opts.Page - 1) * redirect.PageSize
	limit := redirect.PageSize

	q += fmt.Sprintf(" order by dt offset %d limit %d", offset, limit)
	return
}

func (p *Postgres) AgrigatedRedirects(opts redirect.AgrigateOpts) (redirect.Agrigated, error) {
	p.semaphore <- struct{}{}
	defer func() { <-p.semaphore }()
	
	const op = "internal.storage.postgres.agrigatedRedirects"

	wg := sync.WaitGroup{}
	wg.Add(2)
	errC := make(chan error, 2)

	q, args := p.generateAgrigatedReq(opts)
	var res []redirect.Redirect
	var r redirect.Agrigated
	go func() {
		p.semaphore <- struct{}{}
		defer func() { <-p.semaphore }()

		defer wg.Done()

		rows, err := p.db.Master.Query(q, args...)
		if err != nil {
			errC <- err
			return
		}

		defer func() {
			_ = rows.Close()
		}()

		for rows.Next() {
			var tmp redirect.Redirect
			err := rows.Scan(&tmp.ID, &tmp.Alias, &tmp.Date, &tmp.UserAgent)
			if err != nil {
				errC <- err
				return
			}
			res = append(res, tmp)
		}

		r.Alias = opts.Alias
		r.Redirects = res
	}()

	countQ := fmt.Sprintf("select count(*) from %s where alias = $1", RedirectsTable)
	var total int64

	go func() {
		p.semaphore <- struct{}{}
		defer func() { <-p.semaphore }()

		defer wg.Done()

		row := p.db.Master.QueryRow(countQ, opts.Alias)
		if row.Err() != nil {
			errC <- row.Err()
			return
		}
		err := row.Scan(&total)
		if err != nil {
			errC <- err
			return
		}
		r.Total = total
	}()
	wg.Wait()

	if len(errC) != 0 {
		err := <-errC
		return redirect.Agrigated{}, fmt.Errorf("%s: %w", op, err)
	}

	return r, nil
}
