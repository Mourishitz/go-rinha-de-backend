package main

import (
	"context"

	"github.com/redis/go-redis/v9"
)

func (app *Config) WriteToKeyDB(key string, value interface{}) error {
	ctx := context.Background()
	return app.KeyDBClient.Set(ctx, key, value, 0).Err()
}

func (app *Config) ReadFromKeyDB(key string) (string, error) {
	ctx := context.Background()
	value, err := app.KeyDBClient.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return "", nil // Key does not exist
		}
		return "", err // Other error
	}
	return value, nil
}

func (app *Config) ReadAllRequests(service string) (int, error) {
	ctx := context.Background()
	key := service + "_total_requests"
	value, err := app.KeyDBClient.Get(ctx, key).Int()
	if err != nil {
		if err == redis.Nil {
			return 0, nil // Key does not exist
		}
		return 0, err // Other error
	}

	return value, nil
}

func (app *Config) ReadTotalAmount(service string) (float32, error) {
	ctx := context.Background()
	key := service + "_total_amount"
	value, err := app.KeyDBClient.Get(ctx, key).Float32()
	if err != nil {
		if err == redis.Nil {
			return 0, nil // Key does not exist
		}
		return 0, err // Other error
	}

	return value, nil
}
