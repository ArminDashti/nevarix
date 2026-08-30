package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ArminDashti/nevarix-agent/internal/client"
	"github.com/ArminDashti/nevarix-agent/internal/collector"
	"github.com/ArminDashti/nevarix-agent/internal/config"
	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load()
	cfg := config.Load()
	if cfg.AgentToken == "" {
		log.Fatal("NEVARIX_AGENT_TOKEN is required")
	}

	col := collector.New(cfg.AgentID, cfg.Hostname)
	api := client.New(cfg.APIURL, cfg.AgentToken)

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	log.Printf("nevarix-agent starting agentId=%s hostname=%s api=%s interval=%s (timestamps UTC)",
		cfg.AgentID, cfg.Hostname, cfg.APIURL, cfg.CollectInterval)

	ticker := time.NewTicker(cfg.CollectInterval)
	defer ticker.Stop()

	for {
		if err := collectAndSend(ctx, col, api); err != nil {
			log.Printf("send metrics: %v", err)
		}
		select {
		case <-ctx.Done():
			log.Printf("nevarix-agent shutting down")
			return
		case <-ticker.C:
		}
	}
}

func collectAndSend(ctx context.Context, col *collector.Collector, api *client.Client) error {
	snap, err := col.Collect()
	if err != nil {
		return err
	}
	sendCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()
	return api.PostMetrics(sendCtx, snap)
}
