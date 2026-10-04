package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	log.Println("starting telemetry debug lab")

	errCh := make(chan error, 2)
	go func() { errCh <- runConsumer(ctx) }()
	go func() { errCh <- runSensor(ctx) }()

	select {
	case <-ctx.Done():
		log.Println("shutdown requested")
	case err := <-errCh:
		if err != nil {
			log.Fatalf("component stopped: %v", err)
		}
	}
}
