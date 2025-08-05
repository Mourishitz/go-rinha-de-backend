package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/redis/go-redis/v9"
)

type PaymentRequest struct {
	Amount        float64 `json:"amount"`
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

	if resp.StatusCode != http.StatusOK {
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
	err = app.SaveProcessedPayment("default", paymentReq.Amount, paymentReq.RequestedAt)
	FailOnError(err, "Failed to save processed payment in keyDB")
	return true, nil
}

func (app *Config) sendFallbackPayment(body []byte) (bool, error) {
	resp, err := http.Post(app.FallbackServiceURL+"/payments", "application/json", bytes.NewBuffer(body))
	FailOnError(err, "Failed to send payment request to fallback service")

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
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
	err = app.SaveProcessedPayment("fallback", paymentReq.Amount, paymentReq.RequestedAt)
	FailOnError(err, "Failed to save processed payment in keyDB")
	return true, nil
}

func (app *Config) UpdateSummary(amount float64, service string) error {
	app.KeyDBClient.IncrBy(app.Context, service+"_total_requests", 1)
	app.KeyDBClient.IncrByFloat(app.Context, service+"_total_amount", amount)
	return nil
}

func (app *Config) SaveProcessedPayment(service string, amount float64, requestedAt string) error {
	type Payment struct {
		Amount      float64 `json:"amount"`
		RequestedAt string  `json:"requestedAt"`
	}

	t, err := time.Parse(time.RFC3339, requestedAt)
	if err != nil {
		return err
	}

	entry := Payment{
		Amount:      amount,
		RequestedAt: requestedAt,
	}
	jsonValue, err := json.Marshal(entry)
	if err != nil {
		return err
	}

	score := float64(t.UnixMilli())

	return app.KeyDBClient.ZAdd(app.Context, service+"_payments_processed_zset", redis.Z{
		Score:  score,
		Member: jsonValue,
	}).Err()
}
