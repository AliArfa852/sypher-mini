package observability

import "context"

// Tracer provides optional OpenTelemetry-compatible tracing.
// Stub implementation; set Enabled to true and implement spans for full tracing.
type Tracer struct {
	Enabled    bool
	SampleRate float64
}

// NewTracer creates a tracer from config.
func NewTracer(enabled bool, sampleRate float64) *Tracer {
	return &Tracer{Enabled: enabled, SampleRate: sampleRate}
}

// StartSpan starts a span for the given operation. No-op when disabled.
func (t *Tracer) StartSpan(ctx context.Context, name string) (context.Context, func()) {
	if !t.Enabled {
		return ctx, func() {}
	}
	// TODO: OpenTelemetry span
	return ctx, func() {}
}
