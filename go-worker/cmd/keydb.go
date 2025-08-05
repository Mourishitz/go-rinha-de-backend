package main

import (
	"fmt"

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

func (app *Config) NackAndRequeue(stream, group, consumer, messageID string, values map[string]any) error {
	err := app.KeyDBClient.XAck(app.Context, stream, group, messageID).Err()
	if err != nil {
		return fmt.Errorf("failed to ack original message during nack: %w", err)
	}

	_, err = app.KeyDBClient.XAdd(app.Context, &redis.XAddArgs{
		Stream: stream,
		Values: values,
	}).Result()
	if err != nil {
		return fmt.Errorf("failed to requeue message during nack: %w", err)
	}

	return nil
}
