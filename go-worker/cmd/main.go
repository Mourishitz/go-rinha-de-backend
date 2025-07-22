package main

import (
	"log"
	"net/http"
	"os"

	amqp "github.com/rabbitmq/amqp091-go"
)

func FailOnError(err error, msg string) {
	if err != nil {
		log.Panicf("%s: %s", msg, err)
	}
}

type Config struct {
	PaymentServiceURL  string
	FallbackServiceURL string
	DoctorServiceURL   string
	KeyDBServiceURL    string
	IsPaymentsUp       bool
	IsFallbackUp       bool
	HttpClient         *http.Client
}

func main() {
	conn, err := amqp.Dial(os.Getenv("RABBITMQ_URL"))
	FailOnError(err, "Failed to connect to RabbitMQ")
	defer conn.Close()

	ch, err := conn.Channel()
	FailOnError(err, "Failed to open a channel")
	defer ch.Close()

	q, err := ch.QueueDeclare(
		"payments_queue",
		true,
		false,
		false,
		false,
		nil,
	)
	FailOnError(err, "Failed to declare a queue")

	err = ch.Qos(
		1,     // prefetch count
		0,     // prefetch size
		false, // global
	)
	FailOnError(err, "Failed to set QoS")

	msgs, err := ch.Consume(
		q.Name,                   // queue
		os.Getenv("INSTANCE_ID"), // consumer
		false,                    // auto-ack
		false,                    // exclusive
		false,                    // no-local
		false,                    // no-wait
		nil,                      // args
	)
	FailOnError(err, "Failed to register a consumer")

	var forever chan struct{}

	app := Config{
		PaymentServiceURL:  os.Getenv("PAYMENT_PROCESSOR_DEFAULT_URL"),
		FallbackServiceURL: os.Getenv("PAYMENT_PROCESSOR_FALLBACK_URL"),
		DoctorServiceURL:   os.Getenv("DOCTOR_SERVICE_URL"),
		KeyDBServiceURL:    os.Getenv("KEYDB_SERVICE_URL"),
		IsPaymentsUp:       true,
		IsFallbackUp:       true,
		HttpClient:         &http.Client{},
	}

	go func() {
		for d := range msgs {
			_, err = app.SendPayment(d.Body)
			FailOnError(err, "Failed to send payment")
			log.Printf("Done")
			d.Ack(false)
		}
	}()

	log.Printf(" [*] Consuming from " + q.Name + " queue.")
	<-forever
}
