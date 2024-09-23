package proto

import (
	"context"
	"fmt"
	"net"
)

func SocketInit(listenAddr string) error {
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
		go c.Handle()
	}
}
