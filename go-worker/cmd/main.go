package main

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

func FailOnError(err error, msg string) {
	if err != nil {
		log.Printf("%s: %s", msg, err)
	}
}

type Config struct {
	PaymentServiceURL  string
	FallbackServiceURL string
	KeyDBServiceURL    string
	KeyDBClient        *redis.Client
	IsPaymentsUp       bool
	IsFallbackUp       bool
}

func main() {
	ctx := context.Background()

	app := Config{
		PaymentServiceURL:  os.Getenv("PAYMENT_PROCESSOR_DEFAULT_URL"),
		FallbackServiceURL: os.Getenv("PAYMENT_PROCESSOR_FALLBACK_URL"),
		KeyDBServiceURL:    os.Getenv("KEYDB_SERVICE_URL"),
		KeyDBClient: redis.NewClient(&redis.Options{
			Addr: os.Getenv("KEYDB_SERVICE_URL"),
			DB:   0,
		}),
		IsPaymentsUp: true,
		IsFallbackUp: true,
	}

	stream := "payments_stream"
	group := "payments_group"
	consumer := "worker-" + os.Getenv("INSTANCE_ID")

	err := app.KeyDBClient.XGroupCreateMkStream(ctx, stream, group, "$").Err()
	if err != nil && err.Error() != "BUSYGROUP Consumer Group name already exists" {
		log.Fatalf("Failed to create consumer group: %v", err)
	}

	log.Println("[*] Consuming from payments stream using consumer group: ", group)

	for {
		streams, err := app.KeyDBClient.XReadGroup(ctx, &redis.XReadGroupArgs{
			Group:    group,
			Consumer: consumer,
			Streams:  []string{stream, ">"},
			Count:    1,
			Block:    5 * time.Second,
		}).Result()

		if err != nil && err != redis.Nil {
			log.Printf("Error reading from stream: %v", err)
			time.Sleep(1 * time.Second)
			continue
		}

		if len(streams) == 0 || len(streams[0].Messages) == 0 {
			continue
		}

		for _, msg := range streams[0].Messages {
			values := make(map[string]any)
			for k, v := range msg.Values {
				switch k {
				case "amount":
					// Try to convert string to float32
					if strVal, ok := v.(string); ok {
						f, err := strconv.ParseFloat(strVal, 32)
						if err == nil {
							values[k] = float32(f)
							continue
						}
						log.Printf("Invalid float value for 'amount': %v", strVal)
					}
					// fallback if something goes wrong
					values[k] = 0.0
				default:
					values[k] = v
				}
			}
			jsonBytes, err := json.Marshal(values)
			if err != nil {
				log.Printf("Failed to marshal message: %v", err)
				app.KeyDBClient.XAck(ctx, stream, group, msg.ID)
				continue
			}

			if !app.IsPaymentsUp && !app.IsFallbackUp {
				app.NackAndRequeue(ctx, stream, group, consumer, msg.ID, msg.Values)
				app.sendHealthCheckRequest("default")
				app.sendHealthCheckRequest("fallback")
				continue
			}
			shouldAck, err := app.SendPayment(jsonBytes)
			if err != nil {
				FailOnError(err, "Failed to send payment")
			}

			if shouldAck {
				err = app.KeyDBClient.XAck(ctx, stream, group, msg.ID).Err()
				if err != nil {
					log.Printf("Failed to ack message %s: %v", msg.ID, err)
				}
			} else {
				err = app.NackAndRequeue(ctx, stream, group, consumer, msg.ID, msg.Values)
				if err != nil {
					log.Printf("Failed to nack & requeue message %s: %v", msg.ID, err)
				}
			}
		}
	}
}
