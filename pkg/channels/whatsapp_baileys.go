package channels

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/sypherexx/sypher-mini/pkg/bus"
	"github.com/sypherexx/sypher-mini/pkg/extensions"
)

const (
	whatsAppMinInterval = 12 * time.Second
)

// WhatsAppBaileysClient relays outbound messages to the Baileys extension via HTTP.
// Inbound: extension POSTs to gateway /inbound (handled by main).
// Outbound: this client subscribes to msgBus and POSTs to extension /send.
// Per-chat rate limit: min 12s between messages to avoid spam.
type WhatsAppBaileysClient struct {
	baileysURL   string
	msgBus       *bus.MessageBus
	httpClient   *http.Client
	lastSent     map[string]time.Time
	lastSentMu   sync.Mutex
}

// NewWhatsAppBaileysClient creates a Baileys outbound client.
func NewWhatsAppBaileysClient(baileysURL string, msgBus *bus.MessageBus) *WhatsAppBaileysClient {
	url := strings.TrimRight(baileysURL, "/")
	if url == "" {
		url = "http://localhost:3002"
	}
	return &WhatsAppBaileysClient{
		baileysURL: url,
		msgBus:     msgBus,
		httpClient: &http.Client{Timeout: 10 * time.Second},
		lastSent:   make(map[string]time.Time),
	}
}

// Run subscribes to outbound and forwards WhatsApp messages to the extension.
func (w *WhatsAppBaileysClient) Run(ctx context.Context) error {
	for ctx.Err() == nil {
		out, ok := w.msgBus.SubscribeOutbound(ctx)
		if !ok {
			return ctx.Err()
		}
		if out.Channel != "whatsapp" {
			continue
		}
		if err := w.sendWithRateLimit(ctx, out.ChatID, out.Content); err != nil {
			log.Printf("WhatsApp Baileys send error: %v", err)
		}
	}
	return ctx.Err()
}

func (w *WhatsAppBaileysClient) sendWithRateLimit(ctx context.Context, to, content string) error {
	w.lastSentMu.Lock()
	last := w.lastSent[to]
	w.lastSentMu.Unlock()
	if wait := whatsAppMinInterval - time.Since(last); wait > 0 {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(wait):
		}
	}
	if err := w.send(to, content); err != nil {
		return err
	}
	w.lastSentMu.Lock()
	w.lastSent[to] = time.Now()
	w.lastSentMu.Unlock()
	return nil
}

func (w *WhatsAppBaileysClient) send(to, content string) error {
	if to == "" || content == "" {
		return nil
	}
	payload := map[string]string{"to": to, "content": content}
	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	req, err := http.NewRequest("POST", w.baileysURL+"/send", bytes.NewReader(data))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := w.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("extension returned %d", resp.StatusCode)
	}
	return nil
}

// SpawnBaileysExtension starts the whatsapp-baileys extension as a subprocess.
// Returns the process or nil if extension not found or spawn failed.
func SpawnBaileysExtension(baileysURL, coreCallback string) *exec.Cmd {
	wd, _ := os.Getwd()
	exts, err := extensions.DiscoverFromWorkspace(wd)
	if err != nil {
		return nil
	}
	var ext extensions.DiscoveredExtension
	for _, e := range exts {
		if e.Manifest.ID == "whatsapp-baileys" {
			ext = e
			break
		}
	}
	if ext.Dir == "" {
		// Try common paths
		for _, d := range []string{"extensions/whatsapp-baileys", "extensions\\whatsapp-baileys"} {
			if abs, _ := filepath.Abs(d); abs != "" {
				if st, err := os.Stat(abs); err == nil && st.IsDir() {
					ext.Dir = abs
					ext.Manifest = extensions.Manifest{Entry: "dist/index.js", NodeMin: "20"}
					if data, err := os.ReadFile(filepath.Join(abs, "sypher.extension.json")); err == nil {
						_ = json.Unmarshal(data, &ext.Manifest)
					}
					break
				}
			}
		}
	}
	if ext.Dir == "" {
		return nil
	}

	// Check Node version if manifest specifies minimum
	if minVer := ext.Manifest.NodeMin; minVer != "" {
		if !extensions.CheckNodeVersion(minVer) {
			out, _ := exec.Command("node", "-v").Output()
			log.Printf("WhatsApp Baileys requires Node.js %s+. Current: %s Upgrade: https://nodejs.org/", minVer, strings.TrimSpace(string(out)))
			return nil
		}
	}

	port := "3002"
	if u, err := parsePortFromURL(baileysURL); err == nil {
		port = u
	}
	env := append(os.Environ(),
		"SYPHER_CORE_CALLBACK="+coreCallback,
		"PORT="+port,
	)

	entryPath := filepath.Join(ext.Dir, "dist", "index.js")
	if _, err := os.Stat(entryPath); err != nil {
		if ext.Manifest.Setup != "" {
			if !extensions.RunSetup(ext.Dir, ext.Manifest) {
				return nil
			}
		} else {
			nodeModules := filepath.Join(ext.Dir, "node_modules")
			if _, err := os.Stat(nodeModules); os.IsNotExist(err) {
				install := exec.Command("npm", "install")
				install.Dir = ext.Dir
				install.Stdout = os.Stdout
				install.Stderr = os.Stderr
				if install.Run() != nil {
					log.Printf("Baileys extension: npm install failed")
					return nil
				}
			}
			build := exec.Command("npm", "run", "build")
			build.Dir = ext.Dir
			build.Stdout = os.Stdout
			build.Stderr = os.Stderr
			if build.Run() != nil {
				return nil
			}
		}
	}
	if _, err := os.Stat(entryPath); err != nil {
		return nil
	}

	var cmd *exec.Cmd
	if ext.Manifest.Start != "" {
		cmd = extensions.RunStart(ext.Dir, ext.Manifest)
	}
	if cmd == nil {
		cmd = exec.Command("node", entryPath)
	}
	cmd.Dir = ext.Dir
	cmd.Env = env
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Start(); err != nil {
		log.Printf("Failed to spawn Baileys extension: %v", err)
		return nil
	}
	return cmd
}

func parsePortFromURL(urlStr string) (string, error) {
	if strings.HasPrefix(urlStr, "http://") {
		urlStr = urlStr[7:]
	} else if strings.HasPrefix(urlStr, "https://") {
		urlStr = urlStr[8:]
	}
	if idx := strings.LastIndex(urlStr, ":"); idx >= 0 {
		port := urlStr[idx+1:]
		if _, err := strconv.Atoi(port); err == nil {
			return port, nil
		}
	}
	return "", fmt.Errorf("no port")
}
