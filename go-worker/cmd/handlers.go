package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"log"
	"net/http"
)

type PaymentRequest struct {
	Amount        float32 `json:"amount"`
	CorrelationID string  `json:"correlationId"`
	RequestedAt   string  `json:"requestedAt"`
}

func (app *Config) SendPayment(body []byte) (bool, error) {
	if app.IsPaymentsUp {
		return app.sendPaymentRequest(body)
	}

	if app.IsFallbackUp {
		app.CheckServicesStatus("default")
		if app.IsPaymentsUp {
			log.Println("Payments service is back up, retrying payment request")
			return app.sendPaymentRequest(body)
		}
		return app.sendFallbackPayment(body)
	}

	return false, errors.New("both payments and fallback services are down")
}

func (app *Config) sendPaymentRequest(body []byte) (bool, error) {
	resp, err := http.Post(app.PaymentServiceURL+"/payments", "application/json", bytes.NewBuffer(body))
	FailOnError(err, "Failed to send payment request to payments service")

	defer resp.Body.Close()

	if resp.StatusCode == http.StatusInternalServerError {
		log.Printf("Received non-OK response from payments service: %s", resp.Status)
		app.IsPaymentsUp = false
		log.Println("Payments status: ", app.IsPaymentsUp)
		if app.IsFallbackUp {
			log.Println("Attempting to send payment request to fallback service")
			return app.sendFallbackPayment(body)
		}
		return false, errors.New("payments service returned non-OK status")
	}

	// Update summary in keyDB
	var paymentReq PaymentRequest
	err = json.Unmarshal(body, &paymentReq)
	FailOnError(err, "Failed to unmarshal payment request body")

	err = app.UpdateSummary(paymentReq.Amount, "payments")
	FailOnError(err, "Failed to update summary in keyDB")
	return true, nil
}

func (app *Config) sendFallbackPayment(body []byte) (bool, error) {
	resp, err := http.Post(app.FallbackServiceURL+"/payments", "application/json", bytes.NewBuffer(body))
	FailOnError(err, "Failed to send payment request to fallback service")

	defer resp.Body.Close()

	if resp.StatusCode == http.StatusInternalServerError {
		log.Printf("Received non-OK response from fallback service: %s", resp.Status)
		app.IsFallbackUp = false
		log.Println("Fallback status: ", app.IsFallbackUp)
		return false, errors.New("fallback service returned non-OK status, requeuing message")
	}

	// Update summary in keyDB
	var paymentReq PaymentRequest
	err = json.Unmarshal(body, &paymentReq)
	FailOnError(err, "Failed to unmarshal payment request body")

	err = app.UpdateSummary(paymentReq.Amount, "fallback")
	FailOnError(err, "Failed to update summary in keyDB")
	return true, nil
}

func (app *Config) UpdateSummary(amount float32, service string) error {
	totalRequests, err := app.ReadAllRequests(service)
	FailOnError(err, "Failed to read total requests from keyDB")
	totalAmount, err := app.ReadTotalAmount(service)
	FailOnError(err, "Failed to read total amount from keyDB")

	app.WriteToKeyDB(service+"_total_requests", totalRequests+1)
	app.WriteToKeyDB(service+"_total_amount", totalAmount+amount)

	return nil
}
