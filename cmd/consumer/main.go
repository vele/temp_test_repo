package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/sirupsen/logrus"

	"github.com/vele/temp_test_repo/internal/config"
	"github.com/vele/temp_test_repo/internal/event"
	"github.com/vele/temp_test_repo/pkg/logger"
)

func main() {
	cfg := config.Load()
	log := logger.New(logrus.InfoLevel)

	consumer, err := event.NewRabbitConsumer(cfg.RabbitDSN, "user.events", "user.events.console")
	if err != nil {
		log.WithError(err).Fatal("failed to create consumer")
	}
	defer consumer.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		if err := consumer.Consume(ctx, func(ctx context.Context, evt event.Event) error {
			log.WithFields(logrus.Fields{
				"type":   evt.Type,
				"userID": evt.UserID,
			}).Info("event received")
			return nil
		}); err != nil {
			log.WithError(err).Error("consumer stopped")
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop
	cancel()
	log.Info("consumer stopped")
}
