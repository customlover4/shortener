package redis

import (
	"context"
	"errors"

	"github.com/go-redis/redis/v8"
	wbfRedis "github.com/wb-go/wbf/redis"
	"github.com/wb-go/wbf/zlog"
)

type Redis struct {
	rd *wbfRedis.Client
}

func New(addr, password string, db int) *Redis {
	const op = "internal.storage.redis.New"
	r := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})
	cmd := r.Ping(context.Background())
	if cmd.Err() != nil {
		zlog.Logger.Error().AnErr("err", cmd.Err()).Msg(op)
		panic(cmd.Err())
	}
	return &Redis{
		rd: &wbfRedis.Client{
			Client: r,
		},
	}
}

func (r *Redis) Shutdown() {
	const op = "internal.storage.redis.shutdown"

	err := r.rd.Client.Close()
	if err != nil {
		zlog.Logger.Error().AnErr("err", err).Msg(op)
	}
}

func (r *Redis) AddURL(alias, original string) error {
	const op = "internal.storage.redis.AddNotification"

	err := r.rd.Set(context.Background(), alias, original)
	if err != nil {
		zlog.Logger.Error().AnErr("err", err).Msg(op)
		return err
	}

	return nil
}

func (r *Redis) URL(alias string) (string, error) {
	const op = "internal.storage.redis.Get"

	c, err := r.rd.Get(context.Background(), alias)
	if err != nil && !errors.Is(err, redis.Nil) {
		zlog.Logger.Error().AnErr("err", err).Msg(op)
		return "", err
	}
	return string(c), nil
}

func (r *Redis) DeleteNotification(alias string) (int64, error) {
	const op = "internal.storage.redis.DeleteNotification"

	res, err := r.rd.Del(context.Background(), alias).Result()
	if err != nil {
		zlog.Logger.Error().AnErr("err", err).Msg(op)
		return 0, err
	}

	return res, nil
}
