package config

import (
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	APIURL           string
	AgentToken       string
	AgentID          string
	Hostname         string
	CollectInterval  time.Duration
}

func Load() Config {
	hostname := getenv("NEVARIX_HOSTNAME", "")
	if hostname == "" {
		hostname, _ = os.Hostname()
	}
	agentID := getenv("NEVARIX_AGENT_ID", "")
	if agentID == "" {
		agentID = hostname
	}
	return Config{
		APIURL:          strings.TrimRight(getenv("NEVARIX_API_URL", "http://127.0.0.1:8090"), "/"),
		AgentToken:      getenv("NEVARIX_AGENT_TOKEN", ""),
		AgentID:         agentID,
		Hostname:        hostname,
		CollectInterval: time.Duration(getenvInt("COLLECT_INTERVAL_SECONDS", 5)) * time.Second,
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
