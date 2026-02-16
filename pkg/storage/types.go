package storage

import "time"

type Metric struct {
	Timestamp   time.Time `json:"timestamp"`
	DurationNS  int64     `json:"duration_ns"`
	StatusCode  int       `json:"status_code"`
	Success     bool      `json:"success"`
	CertExpiry  time.Time `json:"cert_expiry,omitempty"`
	CertIssuer  string    `json:"cert_issuer,omitempty"`
	CertSubject string    `json:"cert_subject,omitempty"`
}
