package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/manu/octo/pkg/checker"
	"github.com/manu/octo/pkg/storage"
)

type PostgresStorage struct {
	pool *pgxpool.Pool
}

func NewPostgresStorage(ctx context.Context, host, port, user, password, dbName string) (*PostgresStorage, error) {
	connString := fmt.Sprintf("postgres://%s:%s@%s:%s/%s", user, password, host, port, dbName)
	config, err := pgxpool.ParseConfig(connString)
	if err != nil {
		return nil, fmt.Errorf("unable to parse connection string: %w", err)
	}

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("unable to create connection pool: %w", err)
	}

	s := &PostgresStorage{pool: pool}
	if err := s.init(ctx); err != nil {
		pool.Close()
		return nil, err
	}

	return s, nil
}

func (s *PostgresStorage) init(ctx context.Context) error {
	// Create table
	_, err := s.pool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS http_checks (
			time TIMESTAMPTZ NOT NULL,
			endpoint_id TEXT NOT NULL,
			url TEXT NOT NULL,
			method TEXT NOT NULL,
			status_code INTEGER,
			success BOOLEAN,
			duration_ns BIGINT,
			dns_ns BIGINT,
			conn_ns BIGINT,
			tls_ns BIGINT,
			ttfb_ns BIGINT,
			bytes_received BIGINT,
			error TEXT,
			cert_expiry TIMESTAMPTZ,
			cert_issuer TEXT,
			cert_subject TEXT,
			cert_not_before TIMESTAMPTZ,
			cert_not_after TIMESTAMPTZ
		);
	`)
	if err != nil {
		return fmt.Errorf("failed to create table: %w", err)
	}

	// Add columns if they don't exist (migrations)
	migrationQueries := []string{
		"ALTER TABLE http_checks ADD COLUMN IF NOT EXISTS cert_expiry TIMESTAMPTZ",
		"ALTER TABLE http_checks ADD COLUMN IF NOT EXISTS cert_issuer TEXT",
		"ALTER TABLE http_checks ADD COLUMN IF NOT EXISTS cert_subject TEXT",
		"ALTER TABLE http_checks ADD COLUMN IF NOT EXISTS cert_not_before TIMESTAMPTZ",
		"ALTER TABLE http_checks ADD COLUMN IF NOT EXISTS cert_not_after TIMESTAMPTZ",
	}

	for _, query := range migrationQueries {
		if _, err := s.pool.Exec(ctx, query); err != nil {
			return fmt.Errorf("failed to migrate table: %w", err)
		}
	}

	// Convert to hypertable (ignore error if already hypertable)
	// We use a DO block or simple query. TimescaleDB's create_hypertable fails if it already exists unless we handle it.
	// The `if_not_exists => TRUE` parameter is available in recent versions.
	_, err = s.pool.Exec(ctx, `SELECT create_hypertable('http_checks', 'time', if_not_exists => TRUE);`)
	if err != nil {
		// Log warning or inspect error more closely in production
		fmt.Printf("Warning: failed to convert to hypertable (might already exist): %v\n", err)
	}

	return nil
}

func (s *PostgresStorage) Close() {
	s.pool.Close()
}

func (s *PostgresStorage) WriteResult(result checker.Result) error {
	_, err := s.pool.Exec(context.Background(), `
		INSERT INTO http_checks (
			time, endpoint_id, url, method, status_code, success,
			duration_ns, dns_ns, conn_ns, tls_ns, ttfb_ns, bytes_received, error,
			cert_expiry, cert_issuer, cert_subject, cert_not_before, cert_not_after
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18)
	`,
		result.Timestamp,
		result.EndpointID,
		result.URL,
		result.Method,
		result.StatusCode,
		result.Success,
		result.Duration.Nanoseconds(),
		result.DNSDuration.Nanoseconds(),
		result.ConnDuration.Nanoseconds(),
		result.TLSDuration.Nanoseconds(),
		result.TTFB.Nanoseconds(),
		result.BytesReceived,
		result.Error,
		result.CertExpiry,
		result.CertIssuer,
		result.CertSubject,
		result.CertNotBefore,
		result.CertNotAfter,
	)
	return err
}

func (s *PostgresStorage) QueryHistory(ctx context.Context, endpointID string, from, to time.Time) ([]storage.Metric, error) {
	rows, err := s.pool.Query(ctx, `
		SELECT
			time,
			duration_ns,
			status_code,
			success,
			error,
			cert_expiry,
			cert_issuer,
			cert_subject
		FROM http_checks
		WHERE
			endpoint_id = $1
			AND time >= $2
			AND time <= $3
		ORDER BY time ASC
	`, endpointID, from, to)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var metrics []storage.Metric
	for rows.Next() {
		var m storage.Metric
		err := rows.Scan(
			&m.Timestamp, &m.DurationNS, &m.StatusCode, &m.Success, &m.Error,
			&m.CertExpiry, &m.CertIssuer, &m.CertSubject,
		)
		if err != nil {
			return nil, err
		}
		metrics = append(metrics, m)
	}

	return metrics, nil
}
