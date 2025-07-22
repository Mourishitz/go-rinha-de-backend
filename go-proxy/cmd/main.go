package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	amqp "github.com/rabbitmq/amqp091-go"
)

type Config struct {
	paymentsURL  string
	fallbackURL  string
	isPaymentsUp bool
	isFallbackUp bool
	// This can possibly change to Redis in the future
	rabbitMQConn  *amqp.Connection
	rabbitMQChann *amqp.Channel
}

func main() {
	conn, err := amqp.Dial(os.Getenv("RABBITMQ_URL"))
	if err != nil {
		log.Fatalf("Failed to connect to RabbitMQ: %s", err)
	}

	// Open a channel
	ch, err := conn.Channel()
	if err != nil {
		log.Fatalf("Failed to open a channel: %s", err)
	}

	_, err = ch.QueueDeclare(
		"payments_queue",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Fatalf("Failed to declare queue: %s", err)
	}

	defer conn.Close()
	defer ch.Close()

	app := Config{
		isPaymentsUp:  true, // Set to false if the default payment service is down
		isFallbackUp:  true, // Set to false if the fallback service is down
		paymentsURL:   os.Getenv("PROCESSOR_DEFAULT_URL"),
		fallbackURL:   os.Getenv("PROCESSOR_FALLBACK_URL"),
		rabbitMQConn:  conn,
		rabbitMQChann: ch,
	}

	webPort := os.Getenv("APP_PORT")

	log.Printf("Starting API Proxy on port %s\n", webPort)

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%s", webPort),
		Handler: app.routes(),
	}

	err = srv.ListenAndServe()
	if err != nil {
		log.Fatal("Error starting server: ", err)
	}
}
