package utils

import (
	"fmt"
	"net"
	"time"
)

// WaitForService waits for a service to be available on the given host and port
func WaitForService(host string, port int, timeoutSeconds int) error {
	addr := fmt.Sprintf("%s:%d", host, port)
	deadline := time.Now().Add(time.Duration(timeoutSeconds) * time.Second)

	fmt.Printf("Waiting for service at %s...\n", addr)

	for {
		conn, err := net.DialTimeout("tcp", addr, 2*time.Second)
		if err == nil {
			conn.Close()
			fmt.Printf("Service at %s is ready!\n", addr)
			return nil
		}

		if time.Now().After(deadline) {
			return fmt.Errorf("timeout waiting for service at %s", addr)
		}

		time.Sleep(1 * time.Second)
	}
}
