package intent

import (
	"regexp"
	"strings"
)

// Intent represents the classified intent of a message.
type Intent string

const (
	IntentCommand          Intent = "command"
	IntentQuestion         Intent = "question"
	IntentConfigChange     Intent = "config_change"
	IntentAutomationRequest Intent = "automation_request"
	IntentEmergencyAlert   Intent = "emergency_alert"
	IntentChat             Intent = "chat"
)

// Result is the output of the intent parser.
type Result struct {
	Intent Intent
	Params map[string]string
}

// Rule maps a pattern to an intent.
type Rule struct {
	Pattern *regexp.Regexp
	Intent  Intent
}

// Parser classifies message intent before the agent loop.
type Parser struct {
	rules []Rule
}

// New creates a new intent parser with default rules.
func New() *Parser {
	p := &Parser{}
	p.AddDefaultRules()
	return p
}

// AddDefaultRules adds built-in rules for common patterns.
func (p *Parser) AddDefaultRules() {
	// Config commands
	p.AddRule(`^/config\s+`, IntentConfigChange)
	p.AddRule(`^config\s+(get|set)\s+`, IntentConfigChange)

	// Direct command execution (e.g. "run ls -la")
	p.AddRule(`^/run\s+`, IntentCommand)
	p.AddRule(`^run\s+`, IntentCommand)
	p.AddRule(`^!`, IntentCommand) // shell escape

	// Cron/schedule
	p.AddRule(`^/cron\s+`, IntentAutomationRequest)
	p.AddRule(`^schedule\s+`, IntentAutomationRequest)

	// Emergency/alert
	p.AddRule(`^/alert\s+`, IntentEmergencyAlert)
	p.AddRule(`^urgent:`, IntentEmergencyAlert)
}

// AddRule adds a rule. pattern is a regex; intent is the result.
func (p *Parser) AddRule(pattern string, intent Intent) {
	re, err := regexp.Compile(pattern)
	if err != nil {
		return
	}
	p.rules = append(p.rules, Rule{Pattern: re, Intent: intent})
}

// Parse classifies the message. Returns chat for ambiguous/unmatched.
func (p *Parser) Parse(content string) Result {
	content = strings.TrimSpace(content)
	if content == "" {
		return Result{Intent: IntentChat}
	}

	lower := strings.ToLower(content)

	for _, r := range p.rules {
		if r.Pattern.MatchString(lower) {
			return Result{
				Intent: r.Intent,
				Params: extractParams(r.Pattern, content),
			}
		}
	}

	// Default: chat (needs LLM)
	return Result{Intent: IntentChat}
}

func extractParams(re *regexp.Regexp, s string) map[string]string {
	// Simple extraction: could be extended
	return map[string]string{}
}

// NeedsLLM returns true if the intent typically requires the agent loop.
func (r Result) NeedsLLM() bool {
	switch r.Intent {
	case IntentCommand, IntentConfigChange, IntentEmergencyAlert:
		return false
	case IntentQuestion, IntentChat, IntentAutomationRequest:
		return true
	default:
		return true
	}
}
