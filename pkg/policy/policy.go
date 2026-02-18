package policy

import (
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/sypherexx/sypher-mini/pkg/config"
)

// Evaluator evaluates policies for file, network, and rate limits.
type Evaluator struct {
	cfg   *config.Config
	mu    sync.RWMutex
	rates map[string]*rateWindow
}

type rateWindow struct {
	timestamps []time.Time
}

// NewEvaluator creates a policy evaluator.
func NewEvaluator(cfg *config.Config) *Evaluator {
	return &Evaluator{
		cfg:   cfg,
		rates: make(map[string]*rateWindow),
	}
}

// CanAccessFile returns true if the agent can access the path with the given access level.
func (e *Evaluator) CanAccessFile(agentID, path, access string) bool {
	path = config.ExpandPath(path)
	abs, err := filepath.Abs(path)
	if err != nil {
		abs = path
	}

	// Workspace root always allowed for bound agents
	workspace := config.ExpandPath(e.cfg.Agents.Defaults.Workspace)
	wsAbs, _ := filepath.Abs(workspace)
	if strings.HasPrefix(abs, wsAbs) {
		return true
	}

	for _, p := range e.cfg.Policies.Files {
		expanded := config.ExpandPath(strings.TrimSuffix(p.Path, "/**"))
		expAbs, _ := filepath.Abs(expanded)
		matched := abs == expAbs || strings.HasPrefix(abs, expAbs+string(filepath.Separator))
		if !matched {
			if m, _ := filepath.Match(expanded, abs); !m {
				continue
			}
		}
		for _, aid := range p.AgentIDs {
			if aid == "*" || aid == agentID {
				switch p.Access {
				case "read_write":
					return true
				case "read":
					return access == "read"
				case "write":
					return access == "write"
				}
			}
		}
	}
	return false
}

// CanAccessNetwork returns true if the agent can access the given host.
func (e *Evaluator) CanAccessNetwork(agentID, host string) bool {
	// No policies = allow (permissive default)
	if len(e.cfg.Policies.Network) == 0 {
		return true
	}
	for _, n := range e.cfg.Policies.Network {
		agentMatch := false
		for _, aid := range n.AgentIDs {
			if aid == "*" || aid == agentID {
				agentMatch = true
				break
			}
		}
		if !agentMatch {
			continue
		}
		for _, deny := range n.DenyDomains {
			matched, _ := filepath.Match(deny, host)
			if matched {
				return false
			}
		}
		allowed := false
		for _, allow := range n.AllowDomains {
			if allow == "*" {
				allowed = true
				break
			}
			matched, _ := filepath.Match(allow, host)
			if matched {
				allowed = true
				break
			}
		}
		if allowed {
			return true
		}
	}
	return false
}

// CheckRateLimit returns true if the request is within rate limit.
func (e *Evaluator) CheckRateLimit(agentID, toolName string) bool {
	e.mu.Lock()
	defer e.mu.Unlock()

	for _, rl := range e.cfg.Policies.RateLimits {
		if rl.RequestsPerMinute <= 0 {
			continue
		}
		agentMatch := rl.AgentID == "*" || rl.AgentID == agentID
		toolMatch := rl.ToolName == "*" || rl.ToolName == toolName
		if !agentMatch || !toolMatch {
			continue
		}

		key := agentID + ":" + toolName
		if e.rates[key] == nil {
			e.rates[key] = &rateWindow{}
		}
		w := e.rates[key]
		now := time.Now()
		cutoff := now.Add(-time.Minute)
		var valid []time.Time
		for _, ts := range w.timestamps {
			if ts.After(cutoff) {
				valid = append(valid, ts)
			}
		}
		if len(valid) >= rl.RequestsPerMinute {
			return false
		}
		w.timestamps = append(valid, now)
		return true
	}
	return true
}
