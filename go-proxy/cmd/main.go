package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
)

const webPort = "80"

type Config struct {
	paymentsURL  string
	fallbackURL  string
	isPaymentsUp bool
	isFallbackUp bool
}

func main() {
	app := Config{
		isPaymentsUp: true, // Set to false if the default payment service is down
		isFallbackUp: true, // Set to false if the fallback service is down
		paymentsURL:  os.Getenv("PROCESSOR_DEFAULT_URL"),
		fallbackURL:  os.Getenv("PROCESSOR_FALLBACK_URL"),
	}

	log.Printf("Starting API Proxy on port %s\n", webPort)

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%s", webPort),
		Handler: app.routes(),
	}

	err := srv.ListenAndServe()
	if err != nil {
		log.Fatal("Error starting server: ", err)
	}
}
