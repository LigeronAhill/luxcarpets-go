package main

import (
	"context"
	"log/slog"
	"os"
)

func main() {
	ctx := context.Background()
	if err := run(ctx); err != nil {
		slog.Error("Error running server", slog.String("error", err.Error()))
		os.Exit(1)
	}
}
