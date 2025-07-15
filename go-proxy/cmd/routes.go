package main

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

func (app *Config) routes() http.Handler {
	mux := chi.NewRouter()

	mux.Post("/payments", app.Payments)
	mux.Get("/payments-summary", app.PaymentsSummary)

	return mux
}
