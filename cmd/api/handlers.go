package main

import (
	"net/http"
)

func (app *Config) Payments(w http.ResponseWriter, r *http.Request) {
	payload := jsonResponse{
		Error:   false,
		Message: "App is running on the default payment service",
	}

	if !app.isDefaultPaymentUp {
		app.isDefaultPaymentUp = true
		app.writeNoContent(w)
		return
	}
	app.isDefaultPaymentUp = false
	_ = app.writeJSON(w, http.StatusOK, payload)
}
