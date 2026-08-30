# nevarix-api

Host monitoring API for Nevarix: CPU, memory, disk, bandwidth, and Docker.

Agents can push host metrics via `POST /api/v1/ingest/metrics`. All timestamps are **UTC**.

## Stack

- Go + Gin
- PostgreSQL (agent metrics; optional SoftEther history if enabled)
- Docker Engine API

SoftEther is **off by default** (`SOFTETHER_ENABLED=false`). Compose does not run a SoftEther container.

## Quick start (local)

```bash
# Ask before pulling images — then:
docker compose up -d

cp .env.example .env
go run ./cmd/server
```

API listens on `http://127.0.0.1:8090` by default (`/api/v1/...`).

## Agent ingest

Set `AGENT_INGEST_TOKEN` (must match the agent's `NEVARIX_AGENT_TOKEN`).

```http
POST /api/v1/ingest/metrics
Authorization: Bearer <AGENT_INGEST_TOKEN>
Content-Type: application/json
```

When agent data exists, `GET /dashboard`, `/cpu`, `/memory`, `/disk`, and `/bandwidth` prefer the most recently seen agent snapshot; otherwise they fall back to the local collector.

## Environment

See `.env.example`.
