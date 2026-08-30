package store

import (
	"context"
	"embed"
	"time"

	"github.com/ArminDashti/nevarix-api/internal/softether"
	"github.com/jackc/pgx/v5/pgxpool"
)

//go:embed migrations/*.sql
var migrationFS embed.FS

type Store struct {
	pool *pgxpool.Pool
}

type UserStat struct {
	Username             string  `json:"username"`
	ClientIP             string  `json:"clientIp"`
	ASN                  *string `json:"asn"`
	DownloadBytes        uint64  `json:"downloadBytes"`
	UploadBytes          uint64  `json:"uploadBytes"`
	UsageDurationSeconds int64   `json:"usageDurationSeconds"`
}

type SessionLog struct {
	ConnectedAt     time.Time  `json:"connectedAt"`
	DisconnectedAt  *time.Time `json:"disconnectedAt"`
	ClientIP        string     `json:"clientIp"`
	ASN             *string    `json:"asn"`
	DownloadBytes   uint64     `json:"downloadBytes"`
	UploadBytes     uint64     `json:"uploadBytes"`
	DurationSeconds int64      `json:"durationSeconds"`
}

func Connect(ctx context.Context, databaseURL string) (*Store, error) {
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		return nil, err
	}
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, err
	}
	s := &Store{pool: pool}
	if err := s.Migrate(ctx); err != nil {
		pool.Close()
		return nil, err
	}
	return s, nil
}

func (s *Store) Close() {
	if s != nil && s.pool != nil {
		s.pool.Close()
	}
}

func (s *Store) Migrate(ctx context.Context) error {
	for _, name := range []string{
		"migrations/001_softether.sql",
		"migrations/002_agent_metrics.sql",
	} {
		sqlBytes, err := migrationFS.ReadFile(name)
		if err != nil {
			return err
		}
		if _, err := s.pool.Exec(ctx, string(sqlBytes)); err != nil {
			return err
		}
	}
	return nil
}

func (s *Store) SyncOnlineSessions(ctx context.Context, sessions []softether.OnlineSession) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	onlineKeys := make([]string, 0, len(sessions))
	for _, sess := range sessions {
		onlineKeys = append(onlineKeys, sess.SessionKey)
		asn := ""
		if sess.ASN != nil {
			asn = *sess.ASN
		}
		connected := time.Now().UTC()
		if sess.ConnectedAt != nil {
			connected = *sess.ConnectedAt
		}
		_, err := tx.Exec(ctx, `
			INSERT INTO softether_sessions (
				username, client_ip, asn, download_bytes, upload_bytes,
				connected_at, disconnected_at, duration_seconds, session_key
			) VALUES ($1,$2,$3,$4,$5,$6,NULL,$7,$8)
			ON CONFLICT (session_key) DO UPDATE SET
				download_bytes = EXCLUDED.download_bytes,
				upload_bytes = EXCLUDED.upload_bytes,
				duration_seconds = EXCLUDED.duration_seconds,
				client_ip = EXCLUDED.client_ip,
				asn = EXCLUDED.asn
		`, sess.Username, sess.ClientIP, nullIfEmpty(asn), int64(sess.DownloadBytes), int64(sess.UploadBytes),
			connected, sess.SessionDurationSeconds, sess.SessionKey)
		if err != nil {
			return err
		}

		_, err = tx.Exec(ctx, `
			INSERT INTO softether_user_stats (
				username, client_ip, asn, download_bytes, upload_bytes, usage_duration_seconds, updated_at
			) VALUES ($1,$2,$3,$4,$5,$6,NOW())
			ON CONFLICT (username) DO UPDATE SET
				client_ip = EXCLUDED.client_ip,
				asn = COALESCE(EXCLUDED.asn, softether_user_stats.asn),
				download_bytes = GREATEST(softether_user_stats.download_bytes, EXCLUDED.download_bytes),
				upload_bytes = GREATEST(softether_user_stats.upload_bytes, EXCLUDED.upload_bytes),
				usage_duration_seconds = GREATEST(softether_user_stats.usage_duration_seconds, EXCLUDED.usage_duration_seconds),
				updated_at = NOW()
		`, sess.Username, sess.ClientIP, nullIfEmpty(asn), int64(sess.DownloadBytes), int64(sess.UploadBytes), sess.SessionDurationSeconds)
		if err != nil {
			return err
		}
	}

	if len(onlineKeys) == 0 {
		_, err = tx.Exec(ctx, `
			UPDATE softether_sessions
			SET disconnected_at = NOW(),
			    duration_seconds = EXTRACT(EPOCH FROM (NOW() - connected_at))::BIGINT
			WHERE disconnected_at IS NULL
		`)
	} else {
		_, err = tx.Exec(ctx, `
			UPDATE softether_sessions
			SET disconnected_at = NOW(),
			    duration_seconds = EXTRACT(EPOCH FROM (NOW() - connected_at))::BIGINT
			WHERE disconnected_at IS NULL
			  AND NOT (session_key = ANY($1))
		`, onlineKeys)
	}
	if err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (s *Store) ListUserStats(ctx context.Context) ([]UserStat, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT username, COALESCE(client_ip,''), asn, download_bytes, upload_bytes, usage_duration_seconds
		FROM softether_user_stats
		ORDER BY username
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []UserStat{}
	for rows.Next() {
		var u UserStat
		var asn *string
		if err := rows.Scan(&u.Username, &u.ClientIP, &asn, &u.DownloadBytes, &u.UploadBytes, &u.UsageDurationSeconds); err != nil {
			return nil, err
		}
		u.ASN = asn
		out = append(out, u)
	}
	return out, rows.Err()
}

func (s *Store) ListUserSessions(ctx context.Context, username string) ([]SessionLog, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT connected_at, disconnected_at, COALESCE(client_ip,''), asn,
		       download_bytes, upload_bytes, duration_seconds
		FROM softether_sessions
		WHERE username = $1
		ORDER BY connected_at DESC
		LIMIT 200
	`, username)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []SessionLog{}
	for rows.Next() {
		var sLog SessionLog
		var asn *string
		if err := rows.Scan(&sLog.ConnectedAt, &sLog.DisconnectedAt, &sLog.ClientIP, &asn,
			&sLog.DownloadBytes, &sLog.UploadBytes, &sLog.DurationSeconds); err != nil {
			return nil, err
		}
		sLog.ASN = asn
		out = append(out, sLog)
	}
	return out, rows.Err()
}

func nullIfEmpty(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
