package main

import (
	"net/http"

	amqp "github.com/rabbitmq/amqp091-go"
)

func (app *Config) Payments(w http.ResponseWriter, r *http.Request) {
	var requestPayload struct {
		CorrelationID string `json:"correlationId"`
		Amount        string `json:"amount"`
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
			Body:        []byte(`{"correlationId":"` + requestPayload.CorrelationID + `", "amount":"` + requestPayload.Amount + `"}`),
		},
	)
}

func (app *Config) PaymentsSummary(w http.ResponseWriter, r *http.Request) {
	app.writeJSON(w, http.StatusOK, jsonResponse{
		Message: "Payments summary endpoint is up and running",
		Data:    nil,
	})
}
