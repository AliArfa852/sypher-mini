package monitor

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/sypherexx/sypher-mini/pkg/config"
)

// HTTPMonitor polls a URL and triggers alerts on failure.
type HTTPMonitor struct {
	cfg       config.HTTPMonitor
	client    *http.Client
	lastAlert time.Time
	failCount int
	mu        sync.Mutex
	onAlert   func(monitorID, message string)
}

// NewHTTPMonitor creates an HTTP monitor.
func NewHTTPMonitor(cfg config.HTTPMonitor, onAlert func(monitorID, message string)) *HTTPMonitor {
	return &HTTPMonitor{
		cfg:     cfg,
		client:  &http.Client{Timeout: 10 * time.Second},
		onAlert: onAlert,
	}
}

// Run polls the URL at the configured interval until ctx is done.
func (m *HTTPMonitor) Run(ctx context.Context) {
	interval := time.Duration(m.cfg.IntervalSec) * time.Second
	if interval < time.Second {
		interval = time.Minute
	}
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			m.check(ctx)
		}
	}
}

func (m *HTTPMonitor) check(ctx context.Context) {
	req, err := http.NewRequestWithContext(ctx, "GET", m.cfg.URL, nil)
	if err != nil {
		return
	}
	resp, err := m.client.Do(req)
	if err != nil {
		m.recordFailure("request failed: " + err.Error())
		return
	}
	defer resp.Body.Close()

	for _, code := range m.cfg.AlertOnStatus {
		if resp.StatusCode == code {
			m.recordFailure(fmt.Sprintf("HTTP %d", resp.StatusCode))
			return
		}
	}
	m.mu.Lock()
	m.failCount = 0
	m.mu.Unlock()
}

func (m *HTTPMonitor) recordFailure(msg string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.failCount++
	if m.cfg.MinFailures > 0 && m.failCount < m.cfg.MinFailures {
		return
	}
	cooldown := time.Duration(m.cfg.CooldownSec) * time.Second
	if time.Since(m.lastAlert) < cooldown {
		return
	}
	m.lastAlert = time.Now()
	if m.onAlert != nil && m.cfg.AlertViaWhatsApp {
		m.onAlert(m.cfg.ID, msg)
	}
}
