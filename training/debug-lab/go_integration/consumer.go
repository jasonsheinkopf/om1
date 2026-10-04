package main

import (
	"context"
	"log"
	"net"
	"time"
)

func runConsumer(ctx context.Context) error {
	addr, err := net.ResolveUDPAddr("udp", "127.0.0.1:9102")
	if err != nil {
		return err
	}

	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		return err
	}
	defer conn.Close()

	log.Printf("consumer listening on %s", addr)
	buf := make([]byte, 128)

	for {
		if err := conn.SetReadDeadline(time.Now().Add(750 * time.Millisecond)); err != nil {
			return err
		}

		n, remote, err := conn.ReadFromUDP(buf)
		if err != nil {
			if ne, ok := err.(net.Error); ok && ne.Timeout() {
				select {
				case <-ctx.Done():
					return nil
				default:
					log.Println("consumer timeout waiting for telemetry")
					continue
				}
			}
			return err
		}

		log.Printf("consumer received %q from %s", string(buf[:n]), remote)
	}
}
