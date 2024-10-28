package database

import (
	"context"
	"fmt"
	"os"

	"github.com/redis/go-redis/v9"
)

var Ctx = context.Background()

func CreateClient(dbNo int) (*redis.Client, error) {
	if dbNo < 0 {
		return nil, fmt.Errorf("db number cannot be negative")
	}
	if os.Getenv("DB_ADDR") == "" {
		return nil, fmt.Errorf("db address is empty")
	}
	if os.Getenv("DB_PASS") == "" {
		return nil, fmt.Errorf("db pass is empty")
	}
	rdb := redis.NewClient(&redis.Options{
		Addr:     os.Getenv("DB_ADDR"),
		Password: os.Getenv("DB_PASS"),
		DB:       dbNo,
	})
	if rdb == nil {
		return nil, fmt.Errorf("rdb is nil")
	}
	return rdb, nil
}
