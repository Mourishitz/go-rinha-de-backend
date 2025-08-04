package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/redis/go-redis/v9"
)

func (app *Config) Payments(w http.ResponseWriter, r *http.Request) {
	var requestPayload struct {
		CorrelationID string  `json:"correlationId"`
		Amount        float32 `json:"amount"`
	}

	err := app.readJSON(w, r, &requestPayload)
	if err != nil {
		app.errorJSON(w, err, http.StatusBadRequest)
		return
	}

	app.writeNoContent(w)

	go func() {
		ctx := context.Background()

		_, err := app.KeyDBClient.XAdd(ctx, &redis.XAddArgs{
			Stream: "payments_stream",
			Values: map[string]any{
				"correlationId": requestPayload.CorrelationID,
				"amount":        requestPayload.Amount,
				"requestedAt":   time.Now().Format(time.RFC3339Nano),
			},
		}).Result()
		if err != nil {
			log.Printf("Failed to push to payments_stream: %v", err)
		}
	}()
}

func (app *Config) PaymentsSummary(w http.ResponseWriter, r *http.Request) {
	paymentsTotalRequests, err := app.ReadAllRequests("payments")
	FailOnError(err, "Failed to read total requests from KeyDB")

	paymentsTotalAmount, err := app.ReadTotalAmount("payments")
	FailOnError(err, "Failed to read total amount from KeyDB")

	fallbackTotalRequests, err := app.ReadAllRequests("fallback")
	FailOnError(err, "Failed to read fallback total requests from KeyDB")

	fallbackTotalAmount, err := app.ReadTotalAmount("fallback")
	FailOnError(err, "Failed to read fallback total amount from KeyDB")

	app.writeJSON(w, http.StatusOK, map[string]any{
		"default": map[string]any{
			"totalRequests": paymentsTotalRequests,
			"totalAmount":   paymentsTotalAmount,
		},
		"fallback": map[string]any{
			"totalRequests": fallbackTotalRequests,
			"totalAmount":   fallbackTotalAmount,
		},
	})
}
