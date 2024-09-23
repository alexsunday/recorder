package main

import (
	"log/slog"
	"os"
	"recorder/proto"
	"recorder/web"
)

var (
	logger = slog.New(slog.NewTextHandler(os.Stderr, nil).WithGroup("main"))
)

func main() {
	err := web.WebInit("0.0.0.0:18000")
	if err != nil {
		logger.Warn("websocket init failed", "error", err)
		os.Exit(1)
	}

	err = proto.SocketInit("0.0.0.0:12001")
	if err != nil {
		logger.Warn("socket init failed", "error", err)
		os.Exit(1)
	}
}
