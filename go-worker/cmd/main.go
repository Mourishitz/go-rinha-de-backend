package main

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/redis/go-redis/v9"
)

func FailOnError(err error, msg string) {
	if err != nil {
		log.Printf("%s: %s", msg, err)
	}
}

type Config struct {
	Context            context.Context
	PaymentServiceURL  string
	FallbackServiceURL string
	KeyDBServiceURL    string
	KeyDBClient        *redis.Client
	IsPaymentsUp       bool
	IsFallbackUp       bool
}

func main() {
	app := Config{
		Context:            context.Background(),
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

	log.Println("[*] Consuming from payments queue")

	for {
		result, err := app.KeyDBClient.BRPop(app.Context, 0*time.Second, "payments_queue").Result()
		if err != nil {
			continue
		}

		message := result[1]

		if !app.IsPaymentsUp && !app.IsFallbackUp {
			log.Println("All services are down. Requeuing and checking status.")
			_ = app.KeyDBClient.RPush(app.Context, "payments_queue", message).Err()
			app.CheckServicesStatus("default")
			app.CheckServicesStatus("fallback")
			continue
		}

		shouldAck, err := app.SendPayment([]byte(message))
		FailOnError(err, "SendPayment error")

		if !shouldAck {
			log.Println("Message not acknowledged. Requeuing...")
			_ = app.KeyDBClient.RPush(app.Context, "payments_queue", message).Err()
		}
	}
}
