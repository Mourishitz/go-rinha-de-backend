package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"log"
	"net/http"
)

type PaymentRequest struct {
	Amount        float64 `json:"amount"`
	CorrelationID string  `json:"correlationId"`
	RequestedAt   string  `json:"requestedAt"`
}

func (app *Config) SendPayment(body []byte) (any, error) {
	if app.IsPaymentsUp {
		log.Println("Sending payment request to payments service")
		resp, err := http.Post(app.PaymentServiceURL+"/payments", "application/json", bytes.NewBuffer(body))
		FailOnError(err, "Failed to send payment request to payments service")

		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			log.Printf("Received non-OK response from payments service: %s", resp.Status)
			return nil, errors.New("payments service returned non-OK status")
		}

		// Update summary in keyDB
		var paymentReq PaymentRequest
		err = json.Unmarshal(body, &paymentReq)
		FailOnError(err, "Failed to unmarshal payment request body")

		err = app.UpdateSummary(paymentReq.Amount, "payments")
		FailOnError(err, "Failed to update summary in keyDB")

		log.Println("Payment request sent successfully")
		return nil, nil
	}

	if app.IsFallbackUp {
		log.Println("Sending payment request to fallback service")
		resp, err := http.Post(app.FallbackServiceURL+"/payments", "application/json", bytes.NewBuffer(body))
		FailOnError(err, "Failed to send payment request to fallback service")

		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			log.Printf("Received non-OK response from fallback service: %s", resp.Status)
			return nil, errors.New("fallback service returned non-OK status")
		}

		// Update summary in keyDB
		var paymentReq PaymentRequest
		err = json.Unmarshal(body, &paymentReq)
		FailOnError(err, "Failed to unmarshal payment request body")

		err = app.UpdateSummary(paymentReq.Amount, "fallback")
		FailOnError(err, "Failed to update summary in keyDB")

		log.Println("Fallback request sent successfully")
		return nil, nil
	}

	return nil, errors.New("both payments and fallback services are down")
}

func (app *Config) UpdateSummary(amount float64, service string) error {
	totalRequests, err := app.ReadAllRequests(service)
	FailOnError(err, "Failed to read total requests from keyDB")
	totalAmount, err := app.ReadTotalAmount(service)
	FailOnError(err, "Failed to read total amount from keyDB")

	app.WriteToKeyDB(service+"_total_requests", totalRequests+1)
	app.WriteToKeyDB(service+"_total_amount", totalAmount+amount)

	return nil
}
