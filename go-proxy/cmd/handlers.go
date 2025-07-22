package main

import (
	"fmt"
	"net/http"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

func (app *Config) Payments(w http.ResponseWriter, r *http.Request) {
	var requestPayload struct {
		CorrelationID string  `json:"correlationId"`
		Amount        float64 `json:"amount"`
	}

	err := app.readJSON(w, r, &requestPayload)
	if err != nil {
		app.errorJSON(w, err, http.StatusBadRequest)
		return
	}

	app.writeNoContent(w)

	app.rabbitMQChann.Publish(
		"",
		"payments_queue",
		false,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Body: []byte(fmt.Sprintf(
				`{"correlationId":"%s", "amount":%f, "requestedAt":"%s"}`,
				requestPayload.CorrelationID,
				requestPayload.Amount,
				time.Now().Format("2006-01-02T15:04:05.000Z"),
			)),
		},
	)
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
