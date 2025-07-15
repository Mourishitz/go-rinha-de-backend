package main

import (
	"net/http"

	amqp "github.com/rabbitmq/amqp091-go"
)

func (app *Config) Payments(w http.ResponseWriter, r *http.Request) {
	// Publish a message to RabbitMQ
	app.rabbitMQChann.Publish(
		"",
		"payments_queue",
		false,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        []byte(`{"message": "Payment request received"}`),
		},
	)

	app.writeJSON(w, http.StatusOK, jsonResponse{
		Message: "Payments endpoint is up and running, rabbitMQ message sent!",
		Data:    nil,
	})
}

func (app *Config) PaymentsSummary(w http.ResponseWriter, r *http.Request) {
	app.writeJSON(w, http.StatusOK, jsonResponse{
		Message: "Payments summary endpoint is up and running",
		Data:    nil,
	})
}
