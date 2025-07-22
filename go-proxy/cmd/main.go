package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	amqp "github.com/rabbitmq/amqp091-go"
	"github.com/redis/go-redis/v9"
)

type Config struct {
	instance string
	// This can possibly change to Redis in the future
	rabbitMQConn  *amqp.Connection
	rabbitMQChann *amqp.Channel
	KeyDBClient   *redis.Client
}

func FailOnError(err error, msg string) {
	if err != nil {
		log.Panicf("%s: %s", msg, err)
	}
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
		instance:      os.Getenv("INSTANCE_ID"),
		rabbitMQConn:  conn,
		rabbitMQChann: ch,
		KeyDBClient: redis.NewClient(&redis.Options{
			Addr: os.Getenv("KEYDB_SERVICE_URL"),
		}),
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
