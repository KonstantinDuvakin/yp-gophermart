package main

import (
	"context"
	"errors"
	"flag"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/KonstantinDuvakin/yp-gophermart/internal/accrual"
	"github.com/KonstantinDuvakin/yp-gophermart/internal/config"
	"github.com/KonstantinDuvakin/yp-gophermart/internal/handlers/balance"
	"github.com/KonstantinDuvakin/yp-gophermart/internal/handlers/balance/withdraw"
	"github.com/KonstantinDuvakin/yp-gophermart/internal/handlers/login"
	"github.com/KonstantinDuvakin/yp-gophermart/internal/handlers/orders"
	"github.com/KonstantinDuvakin/yp-gophermart/internal/handlers/register"
	"github.com/KonstantinDuvakin/yp-gophermart/internal/handlers/withdrawals"
	"github.com/KonstantinDuvakin/yp-gophermart/internal/middlewares/auth"
	"github.com/KonstantinDuvakin/yp-gophermart/internal/middlewares/gzip"
	auth2 "github.com/KonstantinDuvakin/yp-gophermart/internal/service/auth"
	"github.com/KonstantinDuvakin/yp-gophermart/internal/storage"
	"github.com/go-chi/chi/v5"
)

func main() {
	c := config.NewConfig()
	flag.Parse()
	c.ApplyEnv()

	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		AddSource: true,
	}))
	slog.SetDefault(logger)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	tm := auth2.NewTokenManager(c.JwtSecret)

	store, err := storage.NewPostgresStorage(ctx, c.DatabaseUri)
	if err != nil {
		log.Fatalf("failed to connect to postgres storage: %v", err)
	}
	defer store.Close()

	client := accrual.NewClient(c.AccrualSystemAddress)
	worker := accrual.NewWorker(store, client)

	accrualDone := make(chan struct{})
	go func() {
		worker.Run(ctx)
		close(accrualDone)
	}()

	r := chi.NewRouter()
	r.Route("/api/user", func(r chi.Router) {
		r.Use(gzip.Middleware)

		r.Post("/register", register.PostHandler(store, tm))
		r.Post("/login", login.PostHandler(store, tm))

		r.Group(func(r chi.Router) {
			r.Use(auth.Middleware(tm))

			r.Route("/orders", func(r chi.Router) {
				r.Post("/", orders.PostHandler(store))
				r.Get("/", orders.GetHandler(store))
			})
			r.Route("/balance", func(r chi.Router) {
				r.Get("/", balance.GetHandler(store))
				r.Post("/withdraw", withdraw.PostHandler(store))
			})
			r.Get("/withdrawals", withdrawals.GetHandler(store))
		})

	})

	server := http.Server{
		Addr:    c.RunAddress,
		Handler: r,
	}

	go func() {
		slog.Info("listening", "addr", c.RunAddress)
		if err := server.ListenAndServe(); err != nil {
			if errors.Is(err, http.ErrServerClosed) {
				slog.Info("server stopped")
				return
			}
			log.Fatal(err)
		}
	}()

	<-ctx.Done()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_ = server.Shutdown(shutdownCtx)

	<-accrualDone
}
