// Every Go source file starts with a package declaration.
// "package main" means this file belongs to an executable program rather than a reusable library.
package main

// The import block lists standard-library packages used by this file.
import (
	// context lets us carry cancellation/shutdown signals through multiple functions and goroutines.
	"context"
	// log provides timestamped logging helpers such as Println and Fatalf.
	"log"
	// os gives us operating-system values such as os.Interrupt.
	"os"
	// os/signal lets the program listen for signals sent by the OS or by Ctrl+C.
	"os/signal"
	// syscall exposes lower-level OS signals such as SIGTERM.
	"syscall"
)

// main is the entry point that Go runs when we execute `go run .` or the compiled binary.
func main() {
	// Start with an empty background context, then wrap it in a context that is automatically cancelled
	// when this process receives Ctrl+C (os.Interrupt) or SIGTERM.
	// ctx is passed to the sensor and consumer so both know when the whole program should stop.
	// cancel is a function we can call ourselves if we want to cancel ctx manually.
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)

	// `defer` schedules cancel() to run when main returns.
	// This releases resources associated with the signal-aware context.
	defer cancel()

	// Print a startup message so we know the process reached main successfully.
	log.Println("starting telemetry debug lab")

	// Create a channel that transports error values between goroutines.
	// The channel is buffered with capacity 2 because we launch two components: consumer and sensor.
	// That means each goroutine can send one result without immediately blocking on a receiver.
	errCh := make(chan error, 2)

	// `go` starts this anonymous function in a new goroutine, so it runs concurrently with main.
	// runConsumer(ctx) blocks while listening for UDP telemetry.
	// When it eventually returns, its error (or nil) is sent into errCh.
	go func() { errCh <- runConsumer(ctx) }()

	// Start the simulated sensor in a second goroutine.
	// It also receives the shared ctx so it can stop when shutdown is requested.
	// Its eventual return value is sent into the same errCh channel.
	go func() { errCh <- runSensor(ctx) }()

	// select waits until one of its communication cases becomes ready.
	// In this program we wait for either a shutdown signal or one component to stop.
	select {
	// ctx.Done() is a channel that closes when ctx is cancelled.
	// Receiving from it means Ctrl+C, SIGTERM, or manual cancellation happened.
	case <-ctx.Done():
		// Tell the user that the normal shutdown path was triggered.
		log.Println("shutdown requested")

	// This case waits for either the sensor or consumer to return and send a value into errCh.
	case err := <-errCh:
		// A nil error means the component stopped cleanly; a non-nil error means something failed.
		if err != nil {
			// Fatalf logs the formatted error and then terminates the process with a non-zero exit status.
			// Because Fatalf exits the process immediately, normal deferred cleanup does not run afterward.
			log.Fatalf("component stopped: %v", err)
		}
	}
}
