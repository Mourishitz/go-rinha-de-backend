package main

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

func (app *Config) WriteToKeyDB(key string, value any) error {
	return app.KeyDBClient.Set(app.Context, key, value, 0).Err()
}

func (app *Config) ReadFromKeyDB(key string) (string, error) {
	value, err := app.KeyDBClient.Get(app.Context, key).Result()
	if err != nil {
		if err == redis.Nil {
			return "", nil // Key does not exist
		}
		return "", err // Other error
	}
	return value, nil
}

func (app *Config) ReadAllRequests(service string) (int, error) {
	key := service + "_total_requests"
	value, err := app.KeyDBClient.Get(app.Context, key).Int()
	if err != nil {
		if err == redis.Nil {
			return 0, nil // Key does not exist
		}
		return 0, err // Other error
	}

	return value, nil
}

func (app *Config) ReadTotalAmount(service string) (float64, error) {
	key := service + "_total_amount"
	value, err := app.KeyDBClient.Get(app.Context, key).Float64()
	if err != nil {
		if err == redis.Nil {
			return 0, nil // Key does not exist
		}
		return 0, err // Other error
	}

	return value, nil
}

func (app *Config) ReadRequestsWithParams(service string, from, to time.Time) (int, float64, error) {
	type Payment struct {
		Amount      float64 `json:"amount"`
		RequestedAt string  `json:"requestedAt"`
	}

	results, err := app.KeyDBClient.ZRangeByScore(app.Context, service+"_payments_processed_zset", &redis.ZRangeBy{
		Min: strconv.FormatInt(from.UnixMilli(), 10),
		Max: strconv.FormatInt(to.UnixMilli(), 10),
	}).Result()
	if err != nil {
		return 0, 0, fmt.Errorf("redis query failed: %w", err)
	}

	var totalAmount float64
	var count int

	for _, raw := range results {
		var p Payment
		if err := json.Unmarshal([]byte(raw), &p); err != nil {
			continue // skip malformed entries
		}
		totalAmount += p.Amount
		count++
	}

	return count, totalAmount, nil
}
