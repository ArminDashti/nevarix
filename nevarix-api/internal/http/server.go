package httpserver

import (
	"net/http"
	"strings"
	"time"

	"github.com/ArminDashti/nevarix-api/internal/config"
	"github.com/ArminDashti/nevarix-api/internal/dockerx"
	"github.com/ArminDashti/nevarix-api/internal/metrics"
	"github.com/ArminDashti/nevarix-api/internal/softether"
	"github.com/ArminDashti/nevarix-api/internal/store"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

type Server struct {
	cfg       config.Config
	metrics   *metrics.Collector
	docker    *dockerx.Client
	softether *softether.Client
	store     *store.Store
}

func New(cfg config.Config, m *metrics.Collector, d *dockerx.Client, se *softether.Client, st *store.Store) *Server {
	return &Server{cfg: cfg, metrics: m, docker: d, softether: se, store: st}
}

func (s *Server) Router() *gin.Engine {
	r := gin.New()
	r.Use(gin.Recovery(), gin.Logger())
	r.Use(cors.New(cors.Config{
		AllowOrigins:     s.cfg.CORSOrigins,
		AllowMethods:     []string{"GET", "POST", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	v1 := r.Group("/api/v1")
	{
		v1.GET("/health", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"status": "ok"})
		})
		v1.POST("/ingest/metrics", s.postIngestMetrics)
		v1.GET("/dashboard", s.getDashboard)
		v1.GET("/cpu", s.getCPU)
		v1.GET("/memory", s.getMemory)
		v1.GET("/disk", s.getDisk)
		v1.GET("/bandwidth", s.getBandwidth)
		v1.GET("/docker/images", s.getDockerImages)
		v1.GET("/docker/containers", s.getDockerContainers)
		v1.GET("/softether/sessions", s.getSoftEtherSessions)
		v1.GET("/softether/users", s.getSoftEtherUsers)
		v1.GET("/softether/users/:username/sessions", s.getSoftEtherUserSessions)
	}
	return r
}

func (s *Server) postIngestMetrics(c *gin.Context) {
	if s.cfg.AgentIngestToken == "" || !bearerMatches(c.GetHeader("Authorization"), s.cfg.AgentIngestToken) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	if s.store == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "database unavailable"})
		return
	}
	var payload store.AgentIngestPayload
	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	payload.CollectedAt = payload.CollectedAt.UTC()
	if payload.CollectedAt.IsZero() {
		payload.CollectedAt = time.Now().UTC()
	}
	if err := s.store.UpsertAgentMetrics(c.Request.Context(), payload, s.cfg.MetricsHistoryPoints); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok", "collectedAt": payload.CollectedAt})
}

func bearerMatches(header, token string) bool {
	const prefix = "Bearer "
	if !strings.HasPrefix(header, prefix) {
		return false
	}
	return strings.TrimSpace(strings.TrimPrefix(header, prefix)) == token
}

func (s *Server) agentLatestOrNil(c *gin.Context) *store.AgentLatest {
	if s.store == nil {
		return nil
	}
	latest, err := s.store.GetMostRecentAgentLatest(c.Request.Context())
	if err != nil || latest == nil {
		return nil
	}
	return latest
}

func (s *Server) getDashboard(c *gin.Context) {
	if latest := s.agentLatestOrNil(c); latest != nil {
		diskUsed := uint64(0)
		diskTotal := uint64(0)
		for _, d := range latest.Disks {
			diskUsed += d.UsedBytes
			diskTotal += d.TotalBytes
		}
		c.JSON(http.StatusOK, gin.H{
			"cpuPercent":        latest.CPUPercent,
			"memoryUsedBytes":   latest.MemoryUsedBytes,
			"memoryTotalBytes":  latest.MemoryTotalBytes,
			"memoryUsedPercent": latest.MemoryUsedPercent,
			"diskUsedBytes":     diskUsed,
			"diskTotalBytes":    diskTotal,
			"bandwidthRxBps":    latest.BandwidthRxBps,
			"bandwidthTxBps":    latest.BandwidthTxBps,
			"uptimeSeconds":     latest.UptimeSeconds,
			"collectedAt":       latest.CollectedAt.UTC(),
			"agentId":           latest.AgentID,
			"hostname":          latest.Hostname,
		})
		return
	}

	snap := s.metrics.Snapshot()
	diskUsed := uint64(0)
	diskTotal := uint64(0)
	for _, d := range snap.Disks {
		diskUsed += d.UsedBytes
		diskTotal += d.TotalBytes
	}
	c.JSON(http.StatusOK, gin.H{
		"cpuPercent":        snap.CPUPercent,
		"memoryUsedBytes":   snap.MemoryUsedBytes,
		"memoryTotalBytes":  snap.MemoryTotalBytes,
		"memoryUsedPercent": snap.MemoryUsedPercent,
		"diskUsedBytes":     diskUsed,
		"diskTotalBytes":    diskTotal,
		"bandwidthRxBps":    snap.BandwidthRxBps,
		"bandwidthTxBps":    snap.BandwidthTxBps,
		"uptimeSeconds":     snap.UptimeSeconds,
		"collectedAt":       snap.CollectedAt.UTC(),
	})
}

func (s *Server) getCPU(c *gin.Context) {
	if latest := s.agentLatestOrNil(c); latest != nil {
		series, _ := s.store.ListAgentCPUHistory(c.Request.Context(), latest.AgentID)
		c.JSON(http.StatusOK, gin.H{
			"cpuPercent":  latest.CPUPercent,
			"cores":       latest.Cores,
			"series":      series,
			"collectedAt": latest.CollectedAt.UTC(),
		})
		return
	}
	snap := s.metrics.Snapshot()
	c.JSON(http.StatusOK, gin.H{
		"cpuPercent":  snap.CPUPercent,
		"cores":       snap.Cores,
		"series":      s.metrics.CPUHistory(),
		"collectedAt": snap.CollectedAt.UTC(),
	})
}

func (s *Server) getMemory(c *gin.Context) {
	if latest := s.agentLatestOrNil(c); latest != nil {
		series, _ := s.store.ListAgentMemoryHistory(c.Request.Context(), latest.AgentID)
		c.JSON(http.StatusOK, gin.H{
			"memoryTotalBytes":     latest.MemoryTotalBytes,
			"memoryUsedBytes":      latest.MemoryUsedBytes,
			"memoryAvailableBytes": latest.MemoryAvailableBytes,
			"memoryUsedPercent":    latest.MemoryUsedPercent,
			"series":               series,
			"collectedAt":          latest.CollectedAt.UTC(),
		})
		return
	}
	snap := s.metrics.Snapshot()
	c.JSON(http.StatusOK, gin.H{
		"memoryTotalBytes":     snap.MemoryTotalBytes,
		"memoryUsedBytes":      snap.MemoryUsedBytes,
		"memoryAvailableBytes": snap.MemoryAvailableBytes,
		"memoryUsedPercent":    snap.MemoryUsedPercent,
		"series":               s.metrics.MemoryHistory(),
		"collectedAt":          snap.CollectedAt.UTC(),
	})
}

func (s *Server) getDisk(c *gin.Context) {
	if latest := s.agentLatestOrNil(c); latest != nil {
		c.JSON(http.StatusOK, gin.H{
			"disks":       latest.Disks,
			"collectedAt": latest.CollectedAt.UTC(),
		})
		return
	}
	snap := s.metrics.Snapshot()
	c.JSON(http.StatusOK, gin.H{
		"disks":       snap.Disks,
		"collectedAt": snap.CollectedAt.UTC(),
	})
}

func (s *Server) getBandwidth(c *gin.Context) {
	if latest := s.agentLatestOrNil(c); latest != nil {
		series, _ := s.store.ListAgentBandwidthHistory(c.Request.Context(), latest.AgentID)
		c.JSON(http.StatusOK, gin.H{
			"bandwidthRxBps": latest.BandwidthRxBps,
			"bandwidthTxBps": latest.BandwidthTxBps,
			"series":         series,
			"collectedAt":    latest.CollectedAt.UTC(),
		})
		return
	}
	snap := s.metrics.Snapshot()
	c.JSON(http.StatusOK, gin.H{
		"bandwidthRxBps": snap.BandwidthRxBps,
		"bandwidthTxBps": snap.BandwidthTxBps,
		"series":         s.metrics.BandwidthHistory(),
		"collectedAt":    snap.CollectedAt.UTC(),
	})
}

func (s *Server) getDockerImages(c *gin.Context) {
	if s.docker == nil {
		c.JSON(http.StatusOK, gin.H{"images": []any{}, "error": "docker unavailable"})
		return
	}
	imgs, err := s.docker.ListImages(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"images": []any{}, "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"images": imgs})
}

func (s *Server) getDockerContainers(c *gin.Context) {
	if s.docker == nil {
		c.JSON(http.StatusOK, gin.H{"stacks": []any{}, "error": "docker unavailable"})
		return
	}
	stacks, err := s.docker.ListContainersByStack(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"stacks": []any{}, "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"stacks": stacks})
}

func (s *Server) getSoftEtherSessions(c *gin.Context) {
	sessions, err := s.softether.ListOnlineSessions(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"sessions": []any{}, "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"sessions": sessions})
}

func (s *Server) getSoftEtherUsers(c *gin.Context) {
	if s.store == nil {
		c.JSON(http.StatusOK, gin.H{"users": []any{}, "error": "database unavailable"})
		return
	}
	users, err := s.store.ListUserStats(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"users": []any{}, "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"users": users})
}

func (s *Server) getSoftEtherUserSessions(c *gin.Context) {
	if s.store == nil {
		c.JSON(http.StatusOK, gin.H{"sessions": []any{}, "error": "database unavailable"})
		return
	}
	username := c.Param("username")
	sessions, err := s.store.ListUserSessions(c.Request.Context(), username)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{"sessions": []any{}, "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"username": username, "sessions": sessions})
}
