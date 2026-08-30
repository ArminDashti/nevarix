CREATE TABLE IF NOT EXISTS agent_hosts (
    agent_id     TEXT PRIMARY KEY,
    hostname     TEXT NOT NULL DEFAULT '',
    last_seen_at TIMESTAMPTZ NOT NULL
);

CREATE TABLE IF NOT EXISTS agent_metric_latest (
    agent_id               TEXT PRIMARY KEY REFERENCES agent_hosts(agent_id) ON DELETE CASCADE,
    hostname               TEXT NOT NULL DEFAULT '',
    cpu_percent            DOUBLE PRECISION NOT NULL DEFAULT 0,
    cores_json             JSONB NOT NULL DEFAULT '[]'::jsonb,
    memory_total_bytes     BIGINT NOT NULL DEFAULT 0,
    memory_used_bytes      BIGINT NOT NULL DEFAULT 0,
    memory_available_bytes BIGINT NOT NULL DEFAULT 0,
    memory_used_percent    DOUBLE PRECISION NOT NULL DEFAULT 0,
    disks_json             JSONB NOT NULL DEFAULT '[]'::jsonb,
    bandwidth_rx_bps       DOUBLE PRECISION NOT NULL DEFAULT 0,
    bandwidth_tx_bps       DOUBLE PRECISION NOT NULL DEFAULT 0,
    uptime_seconds         BIGINT NOT NULL DEFAULT 0,
    collected_at           TIMESTAMPTZ NOT NULL
);

CREATE TABLE IF NOT EXISTS agent_metric_points (
    id           BIGSERIAL PRIMARY KEY,
    agent_id     TEXT NOT NULL REFERENCES agent_hosts(agent_id) ON DELETE CASCADE,
    ts           TIMESTAMPTZ NOT NULL,
    cpu_percent  DOUBLE PRECISION NOT NULL DEFAULT 0,
    mem_percent  DOUBLE PRECISION NOT NULL DEFAULT 0,
    rx_bps       DOUBLE PRECISION NOT NULL DEFAULT 0,
    tx_bps       DOUBLE PRECISION NOT NULL DEFAULT 0
);

CREATE INDEX IF NOT EXISTS idx_agent_metric_points_agent_ts
    ON agent_metric_points (agent_id, ts DESC);

CREATE INDEX IF NOT EXISTS idx_agent_hosts_last_seen
    ON agent_hosts (last_seen_at DESC);
