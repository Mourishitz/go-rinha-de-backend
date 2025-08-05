package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"
)

func (app *Config) Payments(w http.ResponseWriter, r *http.Request) {
	var requestPayload struct {
		CorrelationID string  `json:"correlationId"`
		Amount        float64 `json:"amount"`
	}

	err := app.readJSON(w, r, &requestPayload)
	if err != nil {
		app.errorJSON(w, err, http.StatusBadRequest)
		return
	}

	app.writeNoContent(w)

	go func() {
		json, err := json.Marshal(map[string]any{
			"correlationId": requestPayload.CorrelationID,
			"amount":        requestPayload.Amount,
			"requestedAt":   time.Now().Format(time.RFC3339Nano),
		})
		if err != nil {
			log.Fatalf("Failed to marshal payment JSON: %v", err)
		}

		err = app.KeyDBClient.LPush(app.Context, "payments_queue", json).Err()
		if err != nil {
			log.Printf("Failed to push to payments_stream: %v", err)
		}
	}()
}

func (app *Config) PaymentsSummary(w http.ResponseWriter, r *http.Request) {
	params := r.URL.Query()

	if params.Get("from") != "" && params.Get("to") != "" {
		from, _ := time.Parse(time.RFC3339, params.Get("from"))
		to, _ := time.Parse(time.RFC3339, params.Get("to"))

		paymentsTotalRequests, paymentsTotalAmount, err := app.ReadRequestsWithParams("default", from, to)
		FailOnError(err, "Failed to read total requests from KeyDB")

		fallbackTotalRequests, fallbackTotalAmount, err := app.ReadRequestsWithParams("fallback", from, to)
		FailOnError(err, "Failed to read fallback total requests from KeyDB")

		app.writeJSON(w, http.StatusOK, map[string]any{
			"default": map[string]any{
				"totalRequests": paymentsTotalRequests,
				"totalAmount":   paymentsTotalAmount,
			},
			"fallback": map[string]any{
				"totalRequests": fallbackTotalRequests,
				"totalAmount":   fallbackTotalAmount,
			},
		})
		return
	}

	paymentsTotalRequests, err := app.ReadAllRequests("payments")
	FailOnError(err, "Failed to read total requests from KeyDB")

	paymentsTotalAmount, err := app.ReadTotalAmount("payments")
	FailOnError(err, "Failed to read total amount from KeyDB")

	fallbackTotalRequests, err := app.ReadAllRequests("fallback")
	FailOnError(err, "Failed to read fallback total requests from KeyDB")

	fallbackTotalAmount, err := app.ReadTotalAmount("fallback")
	FailOnError(err, "Failed to read fallback total amount from KeyDB")

	app.writeJSON(w, http.StatusOK, map[string]any{
		"default": map[string]any{
			"totalRequests": paymentsTotalRequests,
			"totalAmount":   paymentsTotalAmount,
		},
		"fallback": map[string]any{
			"totalRequests": fallbackTotalRequests,
			"totalAmount":   fallbackTotalAmount,
		},
	})
}
