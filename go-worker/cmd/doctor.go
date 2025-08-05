package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"
)

func (app *Config) CheckServicesStatus(service string) {
	key := service + "_last_check"

	lastCheck, err := app.KeyDBClient.Get(app.Context, key).Time()
	if err != nil {
		log.Printf("Initializing last check time for %s service", service)
		err = app.KeyDBClient.Set(app.Context, key, time.Now(), 0).Err()
		if err != nil {
			log.Printf("Failed to set initial last check time for %s: %v", service, err)
		}
		return
	}

	if time.Since(lastCheck) < 5*time.Second {
		return
	}

	err = app.KeyDBClient.Set(app.Context, key, time.Now(), 0).Err()
	if err != nil {
		log.Printf("Failed to update last check time for %s: %v", service, err)
		return
	}

	app.sendHealthCheckRequest(service)
}

func (app *Config) sendHealthCheckRequest(service string) {
	log.Printf("Sending health check request to %s service", service)
	serviceURL := ""

	switch service {
	case "default":
		serviceURL = app.PaymentServiceURL + "/payments/service-health"
	case "fallback":
		serviceURL = app.FallbackServiceURL + "/payments/service-health"
	}

	resp, err := http.Get(serviceURL)
	if err != nil {
		log.Printf("Failed to send health check request to %s: %v", service, err)
		return
	}
	defer resp.Body.Close()
	var result struct {
		Failing bool `json:"failing"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		log.Printf("Failed to decode response from %s: %v", service, err)
		return
	}

	if result.Failing {
		log.Printf("%s service is failing", service)
		log.Printf("Another check will be performed in 5 seconds")
		app.UpdateLastCheckTime(service)
		return
	}

	log.Printf("%s service is healthy", service)

	switch service {
	case "default":
		app.IsPaymentsUp = true
	case "fallback":
		app.IsFallbackUp = true
	}
	app.UpdateLastCheckTime(service)
}

func (app *Config) UpdateLastCheckTime(service string) {
	key := service + "_last_check"

	_ = app.KeyDBClient.Set(app.Context, key, time.Now(), 0).Err()
}
