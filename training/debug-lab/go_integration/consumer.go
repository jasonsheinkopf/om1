// This file is part of the same executable program as main.go and sensor.go.
package main

// Import the standard-library packages used by the UDP consumer.
import (
	// context gives this function a shared cancellation signal from main.
	"context"
	// log prints status, timeout, and receive messages.
	"log"
	// net provides UDP networking types and functions.
	"net"
	// time is used to create a read timeout so the consumer does not block forever.
	"time"
)

// runConsumer starts a UDP receiver and keeps listening until the context is cancelled or an error occurs.
// The function returns nil for a normal shutdown and a non-nil error for a failure.
func runConsumer(ctx context.Context) error {
	// Convert the text address into Go's structured UDP address type.
	// "udp" selects UDP networking.
	// 127.0.0.1 means "this same computer" (the loopback interface).
	// Port 9102 is where this consumer is intentionally configured to listen.
	// IMPORTANT: the sensor in sensor.go is currently configured for port 9101, so the two sides do NOT match.
	// That mismatch is one of the intentional faults in this debugging exercise.
	addr, err := net.ResolveUDPAddr("udp", "127.0.0.1:9102")

	// If parsing/resolving the address failed, stop immediately and return the error to main.
	if err != nil {
		return err
	}

	// Ask the OS to create a UDP socket and bind it to addr.
	// After this succeeds, packets sent to this computer on UDP port 9102 can be received through conn.
	conn, err := net.ListenUDP("udp", addr)

	// If the socket could not be created or bound, return that error to main.
	if err != nil {
		return err
	}

	// Make sure the UDP socket is closed when runConsumer exits for any normal return path.
	defer conn.Close()

	// Print the address we actually bound to; this is useful evidence when debugging port mismatches.
	log.Printf("consumer listening on %s", addr)

	// Allocate a 128-byte byte slice to use as reusable storage for incoming UDP packet data.
	// UDP gives us raw bytes, so ReadFromUDP needs a byte buffer to write those bytes into.
	buf := make([]byte, 128)

	// Keep receiving packets until we explicitly return from inside the loop.
	for {
		// Set a deadline 750 milliseconds in the future for the NEXT read operation.
		// Without a deadline, ReadFromUDP could block forever if no packet arrives.
		// The timeout lets us periodically wake up, log a useful symptom, and check ctx for shutdown.
		if err := conn.SetReadDeadline(time.Now().Add(750 * time.Millisecond)); err != nil {
			// If setting the socket deadline itself fails, treat that as a component failure.
			return err
		}

		// Wait for one UDP datagram.
		// n is the number of bytes copied into buf.
		// remote is the sender's IP address and port.
		// err is non-nil if the read failed or timed out.
		n, remote, err := conn.ReadFromUDP(buf)

		// Handle any read error before trying to use the received data.
		if err != nil {
			// Try to interpret the generic error as a network error.
			// The `ok` boolean tells us whether that type assertion succeeded.
			if ne, ok := err.(net.Error); ok && ne.Timeout() {
				// A timeout is not automatically fatal; use select to see whether shutdown was requested.
				select {
				// If ctx.Done() is ready, main cancelled the shared context, so exit cleanly.
				case <-ctx.Done():
					return nil

				// default runs immediately when ctx is NOT cancelled.
				default:
					// We timed out while the program is still supposed to be running.
					// In this lab that is a strong clue that telemetry is not reaching this socket.
					log.Println("consumer timeout waiting for telemetry")

					// Skip the rest of this loop iteration and try another read.
					continue
				}
			}

			// Any non-timeout read error is treated as a real failure and returned to main.
			return err
		}

		// The read succeeded.
		// buf[:n] means "only the first n bytes that were actually received," not the unused rest of the buffer.
		// string(...) converts those bytes into printable text.
		// %q prints the payload quoted, and %s prints the sender address.
		log.Printf("consumer received %q from %s", string(buf[:n]), remote)
	}
}
