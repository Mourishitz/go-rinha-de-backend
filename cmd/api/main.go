package main

import (
	"fmt"
	"log"
	"net/http"
)

const webPort = "80"

type Config struct {
	isDefaultPaymentUp bool
}

func main() {
	app := Config{
		isDefaultPaymentUp: true, // Set to false if the default payment service is down
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
