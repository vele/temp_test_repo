package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/sirupsen/logrus"

	"github.com/vele/temp_test_repo/internal/config"
	"github.com/vele/temp_test_repo/internal/event"
	"github.com/vele/temp_test_repo/internal/service"
	postgresstorage "github.com/vele/temp_test_repo/internal/storage/postgres"
	httptransport "github.com/vele/temp_test_repo/internal/transport/http"
	"github.com/vele/temp_test_repo/internal/transport/http/handler"
	"github.com/vele/temp_test_repo/internal/transport/http/middleware"
	"github.com/vele/temp_test_repo/pkg/logger"
)

func main() {
	cfg := config.Load()

	log := logger.New(logrus.InfoLevel)

	repo, err := postgresstorage.NewRepository(cfg.PostgresDSN)
	if err != nil {
		log.WithError(err).Fatal("failed to connect to postgres")
	}

	publisher, err := event.NewRabbitPublisher(cfg.RabbitDSN, "user.events")
	if err != nil {
		log.WithError(err).Fatal("failed to connect to rabbitmq")
	}
	defer publisher.Close()

	userService := service.NewUserService(repo, repo, publisher)
	userHandler := handler.NewUserHandler(userService)
	authHandler := handler.NewAuthHandler(cfg.JWTSecret, cfg.AdminUser, cfg.AdminPassword, cfg.TokenTTL)
	authMiddleware := middleware.NewAuth(cfg.JWTSecret)

	router := httptransport.NewRouter(httptransport.RouterDeps{
		UserHandler: userHandler,
		AuthHandler: authHandler,
		Auth:        authMiddleware,
		Logger:      log,
	})

	server := &http.Server{
		Addr:    cfg.Addr(),
		Handler: router,
	}

	go func() {
		log.Infof("HTTP server listening on %s", cfg.Addr())
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.WithError(err).Fatal("server exited")
		}
	}()

	waitForShutdown(log, server)
}

func waitForShutdown(log *logrus.Logger, server *http.Server) {
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.WithError(err).Error("graceful shutdown failed")
	}
	log.Info("server stopped")
}
