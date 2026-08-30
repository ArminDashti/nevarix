package store

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/ArminDashti/nevarix-api/internal/metrics"
	"github.com/jackc/pgx/v5"
)

type AgentIngestPayload struct {
	AgentID              string               `json:"agentId"`
	Hostname             string               `json:"hostname"`
	CPUPercent           float64              `json:"cpuPercent"`
	Cores                []metrics.CoreSample `json:"cores"`
	MemoryTotalBytes     uint64               `json:"memoryTotalBytes"`
	MemoryUsedBytes      uint64               `json:"memoryUsedBytes"`
	MemoryAvailableBytes uint64               `json:"memoryAvailableBytes"`
	MemoryUsedPercent    float64              `json:"memoryUsedPercent"`
	Disks                []metrics.DiskSample `json:"disks"`
	BandwidthRxBps       float64              `json:"bandwidthRxBps"`
	BandwidthTxBps       float64              `json:"bandwidthTxBps"`
	UptimeSeconds        uint64               `json:"uptimeSeconds"`
	CollectedAt          time.Time            `json:"collectedAt"`
}

type AgentLatest struct {
	AgentID              string
	Hostname             string
	CPUPercent           float64
	Cores                []metrics.CoreSample
	MemoryTotalBytes     uint64
	MemoryUsedBytes      uint64
	MemoryAvailableBytes uint64
	MemoryUsedPercent    float64
	Disks                []metrics.DiskSample
	BandwidthRxBps       float64
	BandwidthTxBps       float64
	UptimeSeconds        uint64
	CollectedAt          time.Time
}

func (s *Store) UpsertAgentMetrics(ctx context.Context, p AgentIngestPayload, historyLimit int) error {
	if s == nil || s.pool == nil {
		return errors.New("database unavailable")
	}
	if p.AgentID == "" {
		return errors.New("agentId is required")
	}
	collectedAt := p.CollectedAt.UTC()
	if collectedAt.IsZero() {
		collectedAt = time.Now().UTC()
	}
	seenAt := time.Now().UTC()
	if historyLimit < 10 {
		historyLimit = 10
	}

	coresJSON, err := json.Marshal(p.Cores)
	if err != nil {
		return err
	}
	if p.Cores == nil {
		coresJSON = []byte("[]")
	}
	disksJSON, err := json.Marshal(p.Disks)
	if err != nil {
		return err
	}
	if p.Disks == nil {
		disksJSON = []byte("[]")
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx, `
		INSERT INTO agent_hosts (agent_id, hostname, last_seen_at)
		VALUES ($1, $2, $3)
		ON CONFLICT (agent_id) DO UPDATE SET
			hostname = EXCLUDED.hostname,
			last_seen_at = EXCLUDED.last_seen_at
	`, p.AgentID, p.Hostname, seenAt)
	if err != nil {
		return err
	}

	_, err = tx.Exec(ctx, `
		INSERT INTO agent_metric_latest (
			agent_id, hostname, cpu_percent, cores_json,
			memory_total_bytes, memory_used_bytes, memory_available_bytes, memory_used_percent,
			disks_json, bandwidth_rx_bps, bandwidth_tx_bps, uptime_seconds, collected_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13)
		ON CONFLICT (agent_id) DO UPDATE SET
			hostname = EXCLUDED.hostname,
			cpu_percent = EXCLUDED.cpu_percent,
			cores_json = EXCLUDED.cores_json,
			memory_total_bytes = EXCLUDED.memory_total_bytes,
			memory_used_bytes = EXCLUDED.memory_used_bytes,
			memory_available_bytes = EXCLUDED.memory_available_bytes,
			memory_used_percent = EXCLUDED.memory_used_percent,
			disks_json = EXCLUDED.disks_json,
			bandwidth_rx_bps = EXCLUDED.bandwidth_rx_bps,
			bandwidth_tx_bps = EXCLUDED.bandwidth_tx_bps,
			uptime_seconds = EXCLUDED.uptime_seconds,
			collected_at = EXCLUDED.collected_at
	`, p.AgentID, p.Hostname, p.CPUPercent, coresJSON,
		int64(p.MemoryTotalBytes), int64(p.MemoryUsedBytes), int64(p.MemoryAvailableBytes), p.MemoryUsedPercent,
		disksJSON, p.BandwidthRxBps, p.BandwidthTxBps, int64(p.UptimeSeconds), collectedAt)
	if err != nil {
		return err
	}

	_, err = tx.Exec(ctx, `
		INSERT INTO agent_metric_points (agent_id, ts, cpu_percent, mem_percent, rx_bps, tx_bps)
		VALUES ($1, $2, $3, $4, $5, $6)
	`, p.AgentID, collectedAt, p.CPUPercent, p.MemoryUsedPercent, p.BandwidthRxBps, p.BandwidthTxBps)
	if err != nil {
		return err
	}

	_, err = tx.Exec(ctx, `
		DELETE FROM agent_metric_points
		WHERE agent_id = $1
		  AND id NOT IN (
			SELECT id FROM agent_metric_points
			WHERE agent_id = $1
			ORDER BY ts DESC
			LIMIT $2
		  )
	`, p.AgentID, historyLimit)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (s *Store) GetMostRecentAgentLatest(ctx context.Context) (*AgentLatest, error) {
	if s == nil || s.pool == nil {
		return nil, nil
	}
	row := s.pool.QueryRow(ctx, `
		SELECT l.agent_id, l.hostname, l.cpu_percent, l.cores_json,
		       l.memory_total_bytes, l.memory_used_bytes, l.memory_available_bytes, l.memory_used_percent,
		       l.disks_json, l.bandwidth_rx_bps, l.bandwidth_tx_bps, l.uptime_seconds, l.collected_at
		FROM agent_metric_latest l
		JOIN agent_hosts h ON h.agent_id = l.agent_id
		ORDER BY h.last_seen_at DESC
		LIMIT 1
	`)
	var a AgentLatest
	var coresRaw, disksRaw []byte
	var memTotal, memUsed, memAvail, uptime int64
	err := row.Scan(
		&a.AgentID, &a.Hostname, &a.CPUPercent, &coresRaw,
		&memTotal, &memUsed, &memAvail, &a.MemoryUsedPercent,
		&disksRaw, &a.BandwidthRxBps, &a.BandwidthTxBps, &uptime, &a.CollectedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	a.MemoryTotalBytes = uint64(memTotal)
	a.MemoryUsedBytes = uint64(memUsed)
	a.MemoryAvailableBytes = uint64(memAvail)
	a.UptimeSeconds = uint64(uptime)
	a.CollectedAt = a.CollectedAt.UTC()
	if len(coresRaw) > 0 {
		_ = json.Unmarshal(coresRaw, &a.Cores)
	}
	if a.Cores == nil {
		a.Cores = []metrics.CoreSample{}
	}
	if len(disksRaw) > 0 {
		_ = json.Unmarshal(disksRaw, &a.Disks)
	}
	if a.Disks == nil {
		a.Disks = []metrics.DiskSample{}
	}
	return &a, nil
}

func (s *Store) ListAgentCPUHistory(ctx context.Context, agentID string) ([]metrics.Point, error) {
	if s == nil || s.pool == nil {
		return nil, nil
	}
	rows, err := s.pool.Query(ctx, `
		SELECT ts, cpu_percent
		FROM agent_metric_points
		WHERE agent_id = $1
		ORDER BY ts ASC
	`, agentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []metrics.Point{}
	for rows.Next() {
		var p metrics.Point
		if err := rows.Scan(&p.Timestamp, &p.Value); err != nil {
			return nil, err
		}
		p.Timestamp = p.Timestamp.UTC()
		out = append(out, p)
	}
	return out, rows.Err()
}

func (s *Store) ListAgentMemoryHistory(ctx context.Context, agentID string) ([]metrics.Point, error) {
	if s == nil || s.pool == nil {
		return nil, nil
	}
	rows, err := s.pool.Query(ctx, `
		SELECT ts, mem_percent
		FROM agent_metric_points
		WHERE agent_id = $1
		ORDER BY ts ASC
	`, agentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []metrics.Point{}
	for rows.Next() {
		var p metrics.Point
		if err := rows.Scan(&p.Timestamp, &p.Value); err != nil {
			return nil, err
		}
		p.Timestamp = p.Timestamp.UTC()
		out = append(out, p)
	}
	return out, rows.Err()
}

func (s *Store) ListAgentBandwidthHistory(ctx context.Context, agentID string) ([]metrics.BandwidthPoint, error) {
	if s == nil || s.pool == nil {
		return nil, nil
	}
	rows, err := s.pool.Query(ctx, `
		SELECT ts, rx_bps, tx_bps
		FROM agent_metric_points
		WHERE agent_id = $1
		ORDER BY ts ASC
	`, agentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []metrics.BandwidthPoint{}
	for rows.Next() {
		var p metrics.BandwidthPoint
		if err := rows.Scan(&p.Timestamp, &p.RxBps, &p.TxBps); err != nil {
			return nil, err
		}
		p.Timestamp = p.Timestamp.UTC()
		out = append(out, p)
	}
	return out, rows.Err()
}
