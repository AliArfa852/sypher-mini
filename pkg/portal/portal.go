package portal

import (
	"embed"
	"encoding/json"
	"io/fs"
	"net/http"
	"strings"

	"github.com/sypherexx/sypher-mini/pkg/agent"
	"github.com/sypherexx/sypher-mini/pkg/bus"
	"github.com/sypherexx/sypher-mini/pkg/config"
	"github.com/sypherexx/sypher-mini/pkg/extensions"
)

//go:embed static/*
var staticFS embed.FS

// Register mounts portal and API routes on mux.
func Register(mux *http.ServeMux, cfg *config.Config, loop *agent.Loop, msgBus *bus.MessageBus) {
	portalMux := http.NewServeMux()

	// API: config (sanitized - mask API keys in response)
	portalMux.HandleFunc("GET /api/config", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		sanitized := sanitizeConfig(cfg)
		_ = json.NewEncoder(w).Encode(sanitized)
	})
	portalMux.HandleFunc("PUT /api/config", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		var updates map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
			http.Error(w, "Bad request", http.StatusBadRequest)
			return
		}
		path := config.GetConfigPath()
		if err := applyConfigUpdates(cfg, updates, path); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]bool{"ok": true})
	})

	// API: tasks
	portalMux.HandleFunc("GET /api/tasks", func(w http.ResponseWriter, r *http.Request) {
		tasks := loop.ListTasks()
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"tasks": tasks})
	})
	portalMux.HandleFunc("GET /api/tasks/{id}", func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		if id == "" {
			http.Error(w, "Missing task ID", http.StatusBadRequest)
			return
		}
		info, ok := loop.GetTask(id)
		if !ok {
			http.Error(w, "Task not found", http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(info)
	})
	portalMux.HandleFunc("POST /api/tasks", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		var req struct {
			Content string `json:"content"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.Content == "" {
			http.Error(w, "Bad request: content required", http.StatusBadRequest)
			return
		}
		msgBus.PublishInbound(bus.InboundMessage{
			Channel:  "portal",
			ChatID:   "portal",
			Content:  req.Content,
			SenderID: "portal",
		})
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"ok": "true", "message": "Task submitted"})
	})
	portalMux.HandleFunc("POST /api/tasks/{id}/cancel", func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		if id == "" {
			http.Error(w, "Missing task ID", http.StatusBadRequest)
			return
		}
		ok := loop.CancelTask(id)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]bool{"ok": ok})
	})

	// API: agents
	portalMux.HandleFunc("GET /api/agents", func(w http.ResponseWriter, r *http.Request) {
		agents := make([]map[string]interface{}, 0, len(cfg.Agents.List))
		for _, a := range cfg.Agents.List {
			agents = append(agents, map[string]interface{}{
				"id":      a.ID,
				"name":    a.Name,
				"default": a.Default,
			})
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"agents": agents})
	})

	// API: extensions
	portalMux.HandleFunc("GET /api/extensions", func(w http.ResponseWriter, r *http.Request) {
		exts, _ := extensions.DiscoverFromWorkspace(".")
		list := make([]map[string]interface{}, 0, len(exts))
		for _, e := range exts {
			list = append(list, map[string]interface{}{
				"id":      e.Manifest.ID,
				"version": e.Manifest.Version,
			})
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"extensions": list})
	})

	// Static portal UI
	subFS, _ := fs.Sub(staticFS, "static")
	portalMux.Handle("GET /portal/", http.StripPrefix("/portal", http.FileServer(http.FS(subFS))))
	portalMux.Handle("GET /portal", http.RedirectHandler("/portal/", http.StatusMovedPermanently))

	mux.Handle("/api/", portalMux)
	mux.Handle("/portal", portalMux)
	mux.Handle("/portal/", portalMux)
}

// sanitizeConfig returns config with API keys masked.
func sanitizeConfig(cfg *config.Config) map[string]interface{} {
	data, _ := json.Marshal(cfg)
	var m map[string]interface{}
	_ = json.Unmarshal(data, &m)
	maskKeys(m, "api_key", "bot_token")
	return m
}

func maskKeys(m map[string]interface{}, keys ...string) {
	for k, v := range m {
		for _, key := range keys {
			if strings.EqualFold(k, key) {
				if s, ok := v.(string); ok && s != "" {
					m[k] = "***"
				}
				break
			}
		}
		if sub, ok := v.(map[string]interface{}); ok {
			maskKeys(sub, keys...)
		}
	}
}

// applyConfigUpdates applies partial updates to config and saves.
func applyConfigUpdates(cfg *config.Config, updates map[string]interface{}, path string) error {
	// Simple merge: update providers.api_key etc.
	if providers, ok := updates["providers"].(map[string]interface{}); ok {
		for name, pc := range providers {
			if p, ok := pc.(map[string]interface{}); ok {
				if key, ok := p["api_key"].(string); ok && key != "" {
					switch name {
					case "cerebras":
						cfg.Providers.Cerebras.APIKey = key
					case "openai":
						cfg.Providers.OpenAI.APIKey = key
					case "anthropic":
						cfg.Providers.Anthropic.APIKey = key
					case "gemini":
						cfg.Providers.Gemini.APIKey = key
					case "deepseek":
						cfg.Providers.DeepSeek.APIKey = key
					}
				}
			}
		}
	}
	return cfg.Save(path)
}
