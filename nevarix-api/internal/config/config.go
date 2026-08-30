package config

import (
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	HTTPAddr             string
	DatabaseURL          string
	CORSOrigins          []string
	SoftEtherContainer   string
	SoftEtherPassword    string
	SoftEtherHub         string
	SoftEtherEnabled     bool
	MetricsHistoryPoints int
	SoftEtherPollEvery   time.Duration
	AgentIngestToken     string
}

func Load() Config {
	return Config{
		HTTPAddr:             getenv("HTTP_ADDR", "127.0.0.1:8090"),
		DatabaseURL:          getenv("DATABASE_URL", "postgres://nevarix:nevarix@127.0.0.1:55432/nevarix?sslmode=disable"),
		CORSOrigins:          splitCSV(getenv("CORS_ORIGINS", "http://127.0.0.1:5173,http://localhost:5173")),
		SoftEtherContainer:   getenv("SOFTETHER_CONTAINER", "nevarix-softether"),
		SoftEtherPassword:    getenv("SOFTETHER_PASSWORD", "nevarix"),
		SoftEtherHub:         getenv("SOFTETHER_HUB", "DEFAULT"),
		SoftEtherEnabled:     getenvBool("SOFTETHER_ENABLED", false),
		MetricsHistoryPoints: getenvInt("METRICS_HISTORY_POINTS", 60),
		SoftEtherPollEvery:   time.Duration(getenvInt("SOFTETHER_POLL_SECONDS", 30)) * time.Second,
		AgentIngestToken:     getenv("AGENT_INGEST_TOKEN", ""),
	}
}

func getenv(key, fallback string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	return fallback
}

func getenvInt(key string, fallback int) int {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return fallback
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return fallback
	}
	return n
}

func getenvBool(key string, fallback bool) bool {
	v := strings.ToLower(strings.TrimSpace(os.Getenv(key)))
	if v == "" {
		return fallback
	}
	return v == "1" || v == "true" || v == "yes"
}

func splitCSV(raw string) []string {
	parts := strings.Split(raw, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			out = append(out, p)
		}
	}
	return out
}
