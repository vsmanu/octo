package checker

import (
	"context"
	"crypto/tls"
	"io"
	"net/http"
	"net/http/httptrace"
	"regexp"
	"strings"
	"time"

	"github.com/manu/octo/pkg/config"
)

// Result holds the metrics of a check
type Result struct {
	Timestamp     time.Time     `json:"timestamp"`
	EndpointID    string        `json:"endpoint_id"`
	SatelliteID   string        `json:"satellite_id"`
	URL           string        `json:"url"`
	Method        string        `json:"method"`
	StatusCode    int           `json:"status_code"`
	Duration      time.Duration `json:"duration"`
	DNSDuration   time.Duration `json:"dns_duration"`
	ConnDuration  time.Duration `json:"conn_duration"`
	TLSDuration   time.Duration `json:"tls_duration"`
	TTFB          time.Duration `json:"ttfb"`
	BytesReceived int64         `json:"bytes_received"`
	Success       bool          `json:"success"`
	Error         string        `json:"error"`

	// SSL/TLS Info
	CertExpiry    time.Time `json:"cert_expiry"`
	CertIssuer    string    `json:"cert_issuer"`
	CertSubject   string    `json:"cert_subject"`
	CertNotBefore time.Time `json:"cert_not_before"`
	CertNotAfter  time.Time `json:"cert_not_after"`
}

// Checker handles the HTTP checks
type Checker struct {
	client *http.Client
}

func NewChecker() *Checker {
	return &Checker{
		client: &http.Client{
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				if len(via) >= 10 {
					return http.ErrUseLastResponse
				}
				return nil
			},
			Transport: &http.Transport{
				MaxIdleConns:        100,
				MaxIdleConnsPerHost: 10,
				IdleConnTimeout:     90 * time.Second,
				TLSClientConfig:     &tls.Config{InsecureSkipVerify: false}, // TODO: make configurable
			},
		},
	}
}

func (c *Checker) Check(ctx context.Context, endpoint config.EndpointConfig) Result {
	result := Result{
		Timestamp:  time.Now(),
		EndpointID: endpoint.ID,
		URL:        endpoint.URL,
		Method:     endpoint.Method,
	}

	var dnsStart, connStart, tlsStart, ttfbStart time.Time

	trace := &httptrace.ClientTrace{
		DNSStart:     func(_ httptrace.DNSStartInfo) { dnsStart = time.Now() },
		DNSDone:      func(_ httptrace.DNSDoneInfo) { result.DNSDuration = time.Since(dnsStart) },
		ConnectStart: func(_, _ string) { connStart = time.Now() },
		ConnectDone: func(_, _ string, err error) {
			if err == nil {
				result.ConnDuration = time.Since(connStart)
			}
		},
		TLSHandshakeStart: func() { tlsStart = time.Now() },
		TLSHandshakeDone: func(_ tls.ConnectionState, err error) {
			if err == nil {
				result.TLSDuration = time.Since(tlsStart)
			}
		},
		GotFirstResponseByte: func() {
			result.TTFB = time.Since(ttfbStart)
		},
	}

	req, err := http.NewRequestWithContext(httptrace.WithClientTrace(ctx, trace), endpoint.Method, endpoint.URL, nil)
	if err != nil {
		result.Error = err.Error()
		return result
	}

	for k, v := range endpoint.Headers {
		req.Header.Add(k, v)
	}

	// Start total timer
	start := time.Now()
	ttfbStart = start // approximate start for TTFB

	resp, err := c.client.Do(req)
	result.Duration = time.Since(start)

	if err != nil {
		result.Error = err.Error()
		return result
	}
	defer resp.Body.Close()

	if resp.TLS != nil && len(resp.TLS.PeerCertificates) > 0 {
		cert := resp.TLS.PeerCertificates[0]
		result.CertExpiry = cert.NotAfter
		result.CertNotAfter = cert.NotAfter
		result.CertNotBefore = cert.NotBefore
		result.CertIssuer = cert.Issuer.String()
		result.CertSubject = cert.Subject.String()
	}

	result.StatusCode = resp.StatusCode

	// Verify status code
	statusOk := false
	if len(endpoint.Validation.StatusCodes) > 0 {
		for _, code := range endpoint.Validation.StatusCodes {
			if code == resp.StatusCode {
				statusOk = true
				break
			}
		}
	} else {
		// Default success codes
		statusOk = resp.StatusCode >= 200 && resp.StatusCode < 300
	}

	if !statusOk {
		result.Error = "status code validation failed"
		return result
	}

	// Verify content if needed
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		result.Error = "failed to read body: " + err.Error()
		return result
	}
	result.BytesReceived = int64(len(bodyBytes))

	if endpoint.Validation.ContentMatch.Pattern != "" {
		bodyStr := string(bodyBytes)
		if endpoint.Validation.ContentMatch.Type == "regex" {
			matched, err := regexp.MatchString(endpoint.Validation.ContentMatch.Pattern, bodyStr)
			if err != nil {
				result.Error = "invalid regex: " + err.Error()
				return result
			}
			if !matched {
				result.Error = "content regex match failed"
				return result
			}
		} else {
			if !strings.Contains(bodyStr, endpoint.Validation.ContentMatch.Pattern) {
				result.Error = "content string match failed"
				return result
			}
		}
	}

	result.Success = true
	return result
}
