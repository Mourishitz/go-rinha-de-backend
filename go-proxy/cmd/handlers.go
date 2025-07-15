package main

import (
	"net/http"
)

func (app *Config) Payments(w http.ResponseWriter, r *http.Request) {
	app.writeJSON(w, http.StatusOK, jsonResponse{
		Message: "Payments endpoint is up and running",
		Data:    nil,
	})
}

func (app *Config) PaymentsSummary(w http.ResponseWriter, r *http.Request) {
	app.writeJSON(w, http.StatusOK, jsonResponse{
		Message: "Payments summary endpoint is up and running",
		Data:    nil,
	})
}
