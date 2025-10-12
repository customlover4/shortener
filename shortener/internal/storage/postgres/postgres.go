package postgres

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/lib/pq"
	_ "github.com/lib/pq"
	"github.com/wb-go/wbf/dbpg"
	"github.com/wb-go/wbf/zlog"
)

const (
	URLTable       = "urls"
	RedirectsTable = "redirects"
)

var (
	ErrNotUnique = errors.New("not unique field value")
)

type Postgres struct {
	db *dbpg.DB

	semaphore chan struct{}
}

func New(host, port, username, password, dbname, sslmode string) *Postgres {
	connStr := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port, username, password, dbname,
	)

	db, err := dbpg.New(connStr, []string{}, &dbpg.Options{
		MaxOpenConns:    100,
		MaxIdleConns:    20,
		ConnMaxLifetime: 2 * time.Hour,
	})
	if err != nil {
		panic(err)
	}
	err = db.Master.Ping()
	if err != nil {
		panic(err)
	}

	p := &Postgres{
		db:        db,
		semaphore: make(chan struct{}, 100),
	}

	return p
}

func (p *Postgres) Shutdown() {
	const op = "internal.storage.redis.shutdown"

	for {
		if len(p.semaphore) == 0 {
			close(p.semaphore)
			break
		}
		time.Sleep(200 * time.Millisecond)
	}

	err := p.db.Master.Close()
	if err != nil {
		zlog.Logger.Error().AnErr("err", err).Msg(op)
	}
}

func (p *Postgres) HandleError(err error) error {
	var pgErr *pq.Error
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case "23505":
			return ErrNotUnique
		}

		if strings.Contains(pgErr.Message, "unique constraint") {
			return ErrNotUnique
		}
	}

	return err
}
