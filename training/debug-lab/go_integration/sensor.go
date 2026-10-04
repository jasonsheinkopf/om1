package main

import (
	"context"
	"fmt"
	"log"
	"math"
	"net"
	"time"
)

func runSensor(ctx context.Context) error {
	addr, err := net.ResolveUDPAddr("udp", "127.0.0.1:9101")
	if err != nil {
		return err
	}

	conn, err := net.DialUDP("udp", nil, addr)
	if err != nil {
		return err
	}
	defer conn.Close()

	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	start := time.Now()
	for {
		select {
		case <-ctx.Done():
			return nil
		case now := <-ticker.C:
			t := now.Sub(start).Seconds()
			value := 12.0 + 2.0*math.Sin(t)
			payload := []byte(fmt.Sprintf("%.2f", value))
			if _, err := conn.Write(payload); err != nil {
				return err
			}
			log.Printf("sensor sent %.2f m/s to %s", value, addr)
		}
	}
}
