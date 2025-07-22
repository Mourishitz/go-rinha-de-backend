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
		Amount        float32 `json:"amount"`
	}

	err := app.readJSON(w, r, &requestPayload)
	if err != nil {
		app.errorJSON(w, err, http.StatusBadRequest)
		return
	}

	app.writeJSON(w, http.StatusOK, jsonResponse{
		Message: "Payments endpoint is up and running, rabbitMQ message sent!",
		Data:    nil,
	})

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
	app.writeJSON(w, http.StatusOK, jsonResponse{
		Message: "Payments summary endpoint is up and running",
		Data:    nil,
	})
}
