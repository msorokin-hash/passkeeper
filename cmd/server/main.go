package main

import (
	"context"
	"fmt"
	"log"
	"os/signal"
	"syscall"
	"time"

	"github.com/msorokin-hash/passkeeper/internal/logger"
	"github.com/msorokin-hash/passkeeper/internal/server"
	"github.com/msorokin-hash/passkeeper/internal/server/config"
	"github.com/sirupsen/logrus"
	"golang.org/x/sync/errgroup"
)

const (
	timeoutServerShutdown = 30 * time.Second
)

func main() {
	if err := run(); err != nil {
		log.Fatalf("service exited with error: %v", err)
	}
}

func run() error {
	rootCtx, stop := signal.NotifyContext(
		context.Background(),
		syscall.SIGTERM,
		syscall.SIGINT,
		syscall.SIGQUIT,
	)
	defer stop()

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	logg, err := logger.NewLogger(cfg.Logging.Level)
	if err != nil {
		return fmt.Errorf("init logger: %w", err)
	}

	srv, err := server.NewServer(rootCtx, cfg, logg)
	if err != nil {
		return fmt.Errorf("init server: %w", err)
	}

	logg.WithFields(logrus.Fields{
		"address":  cfg.Server.Address,
		"logLevel": logg.Level.String(),
	}).Info("starting server")

	g, ctx := errgroup.WithContext(rootCtx)

	g.Go(func() (err error) {
		defer func() {
			if r := recover(); r != nil {
				err = fmt.Errorf("server panic: %v", r)
				logg.WithField("panic", r).Error("panic recovered in server")
			}
		}()

		if err := srv.Run(); err != nil {
			logg.WithError(err).Error("server stopped with error")
			return fmt.Errorf("server run: %w", err)
		}

		logg.Info("server stopped without error")
		return nil
	})

	g.Go(func() error {
		<-ctx.Done()
		logg.Info("shutdown signal received, stopping server...")

		shutdownCtx, cancel := context.WithTimeout(context.Background(), timeoutServerShutdown)
		defer cancel()

		doneCh := make(chan error, 1)
		go func() {
			doneCh <- srv.Stop()
		}()

		select {
		case <-shutdownCtx.Done():
			logg.Error("server stop timed out")
			return fmt.Errorf("server stop timed out: %w", shutdownCtx.Err())
		case err := <-doneCh:
			if err != nil {
				logg.WithError(err).Error("error stopping server")
				return fmt.Errorf("stop server: %w", err)
			}
			logg.Info("server fully stopped")
			return nil
		}
	})

	if err := g.Wait(); err != nil {
		return err
	}

	logg.Info("service has been shutdown cleanly")
	return nil
}
