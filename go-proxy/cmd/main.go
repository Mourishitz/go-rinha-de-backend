package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/redis/go-redis/v9"
)

type Config struct {
	Context     context.Context
	instance    string
	KeyDBClient *redis.Client
}

func FailOnError(err error, msg string) {
	if err != nil {
		log.Panicf("%s: %s", msg, err)
	}
}

func main() {
	app := Config{
		Context:  context.Background(),
		instance: os.Getenv("INSTANCE_ID"),
		KeyDBClient: redis.NewClient(&redis.Options{
			Addr: os.Getenv("KEYDB_SERVICE_URL"),
		}),
	}

	webPort := os.Getenv("APP_PORT")

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
