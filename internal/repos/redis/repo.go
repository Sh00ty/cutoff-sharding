package redis

import (
	"context"
	"strconv"
	"time"

	"github.com/Sh00ty/cutoff-sharding/pkg/id/snowflake"
	redis "github.com/redis/go-redis/v9"
)

type Repository struct {
	clnt redis.UniversalClient
}

func New(ctx context.Context, addr string) (*Repository, error) {
	clnt := redis.NewClient(&redis.Options{
		Addr:       addr,
		ClientName: "cutoffs",
	})
	if err := clnt.Ping(ctx).Err(); err != nil {
		return nil, err
	}
	return &Repository{clnt: clnt}, nil
}

func (r *Repository) Get(ctx context.Context, id snowflake.ID) (*int64, error) {
	res := r.clnt.Get(ctx, strconv.FormatUint(uint64(id), 10))
	if res.Err() != nil {
		return nil, res.Err()
	}

	resInt, err := res.Int64()
	return &resInt, err
}
func (r *Repository) Set(ctx context.Context, id snowflake.ID, val int64) error {
	return r.clnt.Set(ctx, strconv.FormatUint(uint64(id), 10), val, time.Hour).Err()
}
func (r *Repository) Remove(ctx context.Context, id snowflake.ID) error {
	return r.clnt.Del(ctx, strconv.FormatUint(uint64(id), 10)).Err()
}
