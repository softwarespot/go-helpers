package helpers

import (
	"context"
	"os"
	"os/signal"
)

// SignalTrap sets up a signal handler that listens for specified OS signals.
// When any of the specified signals are received, it cancels the provided context
func SignalTrap(ctx context.Context, sig ...os.Signal) (context.Context, context.CancelFunc) {
	ctx, cancel := context.WithCancel(ctx)
	go func() {
		signals := make(chan os.Signal, 1)
		signal.Notify(signals, sig...)
		select {
		case <-ctx.Done():
		case <-signals:
		}
		cancel()
	}()
	return ctx, cancel
}
