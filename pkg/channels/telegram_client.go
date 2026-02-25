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
	"strings"
	"sync"
	"time"

	"github.com/sypherexx/sypher-mini/pkg/bus"
	"github.com/sypherexx/sypher-mini/pkg/config"
	"github.com/sypherexx/sypher-mini/pkg/extensions"
	"github.com/sypherexx/sypher-mini/pkg/utils"
)

const (
	defaultTelegramMinInterval = 12 * time.Second
)

// TelegramClient relays outbound messages to the Telegram extension via HTTP.
type TelegramClient struct {
	botURL      string
	msgBus      *bus.MessageBus
	httpClient  *http.Client
	lastSent    map[string]time.Time
	lastSentMu  sync.Mutex
	minInterval time.Duration
}

// NewTelegramClient creates a Telegram outbound client.
func NewTelegramClient(botURL string, msgBus *bus.MessageBus, minIntervalSec int) *TelegramClient {
	url := strings.TrimRight(botURL, "/")
	if url == "" {
		url = "http://localhost:3003"
	}
	interval := defaultTelegramMinInterval
	if minIntervalSec > 0 {
		interval = time.Duration(minIntervalSec) * time.Second
	}
	return &TelegramClient{
		botURL:      url,
		msgBus:      msgBus,
		httpClient:  &http.Client{Timeout: 10 * time.Second},
		lastSent:    make(map[string]time.Time),
		minInterval: interval,
	}
}

// Run subscribes to outbound and forwards Telegram messages to the extension.
func (t *TelegramClient) Run(ctx context.Context) error {
	for ctx.Err() == nil {
		out, ok := t.msgBus.SubscribeOutbound(ctx)
		if !ok {
			return ctx.Err()
		}
		if out.Channel != "telegram" {
			continue
		}
		log.Printf("[gateway] telegram outbound to=%q content=%q", out.ChatID, utils.Truncate(out.Content, 60))
		if err := t.sendWithRateLimit(ctx, out.ChatID, out.Content); err != nil {
			log.Printf("Telegram send error: %v", err)
		}
	}
	return ctx.Err()
}

func (t *TelegramClient) sendWithRateLimit(ctx context.Context, to, content string) error {
	t.lastSentMu.Lock()
	last := t.lastSent[to]
	t.lastSentMu.Unlock()
	if wait := t.minInterval - time.Since(last); wait > 0 {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(wait):
		}
	}
	if err := t.send(to, content); err != nil {
		return err
	}
	t.lastSentMu.Lock()
	t.lastSent[to] = time.Now()
	t.lastSentMu.Unlock()
	return nil
}

func (t *TelegramClient) send(to, content string) error {
	if to == "" || content == "" {
		return nil
	}
	payload := map[string]string{"to": to, "content": content}
	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	req, err := http.NewRequest("POST", t.botURL+"/send", bytes.NewReader(data))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := t.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("extension returned %d", resp.StatusCode)
	}
	return nil
}

// SpawnTelegramExtension starts the telegram-bot extension as a subprocess.
func SpawnTelegramExtension(cfg *config.TelegramConfig, coreCallback string) *exec.Cmd {
	if !cfg.Enabled || cfg.BotURL == "" {
		return nil
	}
	wd, _ := os.Getwd()
	exts, err := extensions.DiscoverFromWorkspace(wd)
	if err != nil {
		return nil
	}
	var ext extensions.DiscoveredExtension
	for _, e := range exts {
		if e.Manifest.ID == "telegram-bot" {
			ext = e
			break
		}
	}
	if ext.Dir == "" {
		fallbacks := []string{"extensions/telegram-bot", "extensions\\telegram-bot"}
		if home, err := os.UserHomeDir(); err == nil {
			fallbacks = append(fallbacks, filepath.Join(home, "sypher-mini", "extensions", "telegram-bot"))
		}
		for _, d := range fallbacks {
			abs := d
			if !filepath.IsAbs(d) {
				abs, _ = filepath.Abs(d)
			}
			if abs != "" {
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
	if minVer := ext.Manifest.NodeMin; minVer != "" {
		if !extensions.CheckNodeVersion(minVer) {
			out, _ := exec.Command("node", "-v").Output()
			log.Printf("Telegram extension requires Node.js %s+. Current: %s", minVer, strings.TrimSpace(string(out)))
			return nil
		}
	}
	port := "3003"
	if u, err := parsePortFromURL(cfg.BotURL); err == nil {
		port = u
	}
	env := append(os.Environ(),
		"SYPHER_CORE_CALLBACK="+coreCallback,
		"PORT="+port,
	)
	if cfg.BotToken != "" {
		env = append(env, "TELEGRAM_BOT_TOKEN="+cfg.BotToken)
	}
	entryPath := filepath.Join(ext.Dir, "dist", "index.js")
	if _, err := os.Stat(entryPath); err != nil {
		if ext.Manifest.Setup != "" {
			if !extensions.RunSetup(ext.Dir, ext.Manifest) {
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
		log.Printf("Failed to spawn Telegram extension: %v", err)
		return nil
	}
	return cmd
}

