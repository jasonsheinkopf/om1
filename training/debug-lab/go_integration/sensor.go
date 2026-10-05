// This file is part of the same executable program as main.go and consumer.go.
package main

// Import the standard-library packages used by the simulated UDP sensor.
import (
	// context lets the sensor stop when main cancels the shared program context.
	"context"
	// fmt formats the floating-point sensor value as text before we send it.
	"fmt"
	// log prints each transmitted value so we can observe what the sensor believes it is doing.
	"log"
	// math gives us math.Sin, which we use to make the fake speed rise and fall smoothly.
	"math"
	// net provides UDP networking functions and types.
	"net"
	// time provides the periodic ticker and elapsed-time calculations.
	"time"
)

// runSensor simulates a sensor that sends one changing telemetry value every 500 milliseconds.
// It returns nil on a normal context-driven shutdown or an error if networking fails.
func runSensor(ctx context.Context) error {
	// Convert the destination text address into Go's UDP address type.
	// 127.0.0.1 means the packet never leaves this computer.
	// Port 9101 is the destination port this sensor is intentionally configured to send to.
	// IMPORTANT: consumer.go listens on port 9102, so these ports currently do NOT match.
	// That intentional mismatch is the bug this integration lab is designed to expose.
	addr, err := net.ResolveUDPAddr("udp", "127.0.0.1:9101")

	// If the destination address cannot be resolved, stop and report the error to main.
	if err != nil {
		return err
	}

	// Create a UDP connection whose remote destination is addr.
	// The nil local address tells Go/the OS to choose an appropriate local IP and ephemeral source port automatically.
	// UDP is connectionless at the protocol level, but DialUDP gives us a convenient conn object with a fixed destination.
	conn, err := net.DialUDP("udp", nil, addr)

	// If the UDP socket cannot be created, return the error to main.
	if err != nil {
		return err
	}

	// Close the UDP socket automatically when runSensor returns normally.
	defer conn.Close()

	// Create a ticker that sends the current time on ticker.C every 500 milliseconds.
	// This is what controls the simulated sensor's 2 Hz update rate.
	ticker := time.NewTicker(500 * time.Millisecond)

	// Always stop the ticker when this function exits so its internal timer resources are released.
	defer ticker.Stop()

	// Remember when the sensor started so we can compute elapsed time for the sine wave below.
	start := time.Now()

	// Keep running until either the context is cancelled or a write fails.
	for {
		// Wait for whichever happens first: shutdown or the next periodic ticker event.
		select {
		// ctx.Done() becomes ready when main receives Ctrl+C/SIGTERM and cancels the shared context.
		case <-ctx.Done():
			// Returning nil says this was an expected, clean shutdown rather than a failure.
			return nil

		// Every 500 ms, ticker.C provides the timestamp for that tick and stores it in the local variable now.
		case now := <-ticker.C:
			// Subtract the start time from the current tick time to get elapsed duration,
			// then convert that duration into floating-point seconds.
			t := now.Sub(start).Seconds()

			// Generate a fake wheel-speed-like signal.
			// The baseline is 12.0 m/s and the sine term makes it smoothly vary by plus/minus 2.0 m/s.
			value := 12.0 + 2.0*math.Sin(t)

			// Format the float with two decimal places, then convert the resulting string into raw bytes.
			// UDP transmits bytes, not Go float64 values directly.
			payload := []byte(fmt.Sprintf("%.2f", value))

			// Write the payload to the destination associated with conn (127.0.0.1:9101).
			// We ignore the byte count with `_` but keep and test the returned error.
			if _, err := conn.Write(payload); err != nil {
				// A failed write stops the sensor and reports the failure back to main through errCh.
				return err
			}

			// Log what we attempted to send and where we sent it.
			// This log can be compared with the consumer's listening address during debugging.
			log.Printf("sensor sent %.2f m/s to %s", value, addr)
		}
	}
}
