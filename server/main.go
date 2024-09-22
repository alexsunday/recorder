package main

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"os"
)

var (
	logger = slog.New(slog.NewTextHandler(os.Stderr, nil))
)

func main() {
	err := listenSocket("0.0.0.0:12001")
	if err != nil {
		logger.Warn("listen socket failed", "error", err)
		os.Exit(1)
	}
}

func listenSocket(listenAddr string) error {
	lAddr, err := net.ResolveTCPAddr("tcp", listenAddr)
	if err != nil {
		return fmt.Errorf("resolve addr failed %w", err)
	}
	listener, err := net.ListenTCP("tcp", lAddr)
	if err != nil {
		return fmt.Errorf("listen tcp %s failed %w", listenAddr, err)
	}

	ctx := context.Background()
	for {
		conn, err := listener.AcceptTCP()
		if err != nil {
			return fmt.Errorf("accept tcp failed %w", err)
		}

		c := NewConnection(ctx, conn)
		go c.ReadLoop()
	}
}
