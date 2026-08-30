package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ArminDashti/nevarix-api/internal/asn"
	"github.com/ArminDashti/nevarix-api/internal/config"
	"github.com/ArminDashti/nevarix-api/internal/dockerx"
	httpserver "github.com/ArminDashti/nevarix-api/internal/http"
	"github.com/ArminDashti/nevarix-api/internal/metrics"
	"github.com/ArminDashti/nevarix-api/internal/softether"
	"github.com/ArminDashti/nevarix-api/internal/store"
	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load()
	cfg := config.Load()

	collector := metrics.NewCollector(cfg.MetricsHistoryPoints)

	var dockerClient *dockerx.Client
	if dc, err := dockerx.New(); err != nil {
		log.Printf("docker client unavailable: %v", err)
	} else {
		dockerClient = dc
		defer dockerClient.Close()
	}

	resolver := asn.NewResolver()
	seClient := softether.New(cfg.SoftEtherContainer, cfg.SoftEtherPassword, cfg.SoftEtherHub, cfg.SoftEtherEnabled, resolver)

	ctx := context.Background()
	var db *store.Store
	if st, err := store.Connect(ctx, cfg.DatabaseURL); err != nil {
		log.Printf("postgres unavailable (SoftEther history disabled): %v", err)
	} else {
		db = st
		defer db.Close()
		if cfg.SoftEtherEnabled {
			go pollSoftEther(cfg, seClient, db)
		}
	}

	srv := httpserver.New(cfg, collector, dockerClient, seClient, db)
	httpServer := &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           srv.Router(),
		ReadHeaderTimeout: 10 * time.Second,
	}

	go func() {
		log.Printf("nevarix-api listening on http://%s", cfg.HTTPAddr)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server error: %v", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = httpServer.Shutdown(shutdownCtx)
}

func pollSoftEther(cfg config.Config, se *softether.Client, db *store.Store) {
	ticker := time.NewTicker(cfg.SoftEtherPollEvery)
	defer ticker.Stop()
	for {
		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		sessions, err := se.ListOnlineSessions(ctx)
		if err != nil {
			log.Printf("softether poll: %v", err)
		} else if err := db.SyncOnlineSessions(ctx, sessions); err != nil {
			log.Printf("softether sync: %v", err)
		}
		cancel()
		<-ticker.C
	}
}
