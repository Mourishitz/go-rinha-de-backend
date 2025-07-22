package main

import (
	"bytes"
	"errors"
	"log"
	"net/http"
)

func (app *Config) SendPayment(body []byte) (any, error) {
	if app.IsPaymentsUp {
		log.Println("Sending payment request to payments service")
		resp, err := http.Post(app.PaymentServiceURL+"/payments", "application/json", bytes.NewBuffer(body))

		FailOnError(err, "Failed to send payment request to payments service")
		defer resp.Body.Close()
		log.Println("Payment request sent successfully")
		log.Println("Response status:", resp.Status)
		log.Println("Response body:", resp.Body)
		log.Println("Called URL:", app.PaymentServiceURL+"/payments")
		log.Println("Request body:", string(body))
		return nil, nil
	}

	if app.IsFallbackUp {
		log.Println("Payments service is down, sending request to fallback service")
		resp, err := http.Post(app.FallbackServiceURL+"/payments", "application/json", bytes.NewBuffer(body))

		FailOnError(err, "Failed to send payment request to fallback service")
		defer resp.Body.Close()
		log.Println("Payment request sent successfully")
		log.Println("Response status:", resp.Status)
		return nil, nil
	}

	return nil, errors.New("both payments and fallback services are down")
}
