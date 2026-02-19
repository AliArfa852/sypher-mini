package utils

import (
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
)

// IsAudioFile checks if a file is an audio file based on its filename extension and content type.
func IsAudioFile(filename, contentType string) bool {
	audioExtensions := []string{".mp3", ".wav", ".ogg", ".m4a", ".flac", ".aac", ".wma"}
	audioTypes := []string{"audio/", "application/ogg", "application/x-ogg"}

	for _, ext := range audioExtensions {
		if strings.HasSuffix(strings.ToLower(filename), ext) {
			return true
		}
	}

	for _, audioType := range audioTypes {
		if strings.HasPrefix(strings.ToLower(contentType), audioType) {
			return true
		}
	}

	return false
}

// SanitizeFilename removes potentially dangerous characters from a filename.
func SanitizeFilename(filename string) string {
	base := filepath.Base(filename)
	base = strings.ReplaceAll(base, "..", "")
	base = strings.ReplaceAll(base, "/", "_")
	base = strings.ReplaceAll(base, "\\", "_")
	return base
}

// DownloadOptions holds optional parameters for downloading files.
type DownloadOptions struct {
	Timeout      time.Duration
	ExtraHeaders map[string]string
}

// DownloadFile downloads a file from URL to a local temp directory.
// Returns the local file path or empty string on error.
func DownloadFile(url, filename string, opts DownloadOptions) string {
	if opts.Timeout == 0 {
		opts.Timeout = 60 * time.Second
	}

	mediaDir := filepath.Join(os.TempDir(), "sypher-mini_media")
	if err := os.MkdirAll(mediaDir, 0700); err != nil {
		return ""
	}

	safeName := SanitizeFilename(filename)
	localPath := filepath.Join(mediaDir, uuid.New().String()[:8]+"_"+safeName)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return ""
	}

	for key, value := range opts.ExtraHeaders {
		req.Header.Set(key, value)
	}

	client := &http.Client{Timeout: opts.Timeout}
	resp, err := client.Do(req)
	if err != nil {
		return ""
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return ""
	}

	out, err := os.Create(localPath)
	if err != nil {
		return ""
	}
	defer out.Close()

	if _, err := io.Copy(out, resp.Body); err != nil {
		out.Close()
		os.Remove(localPath)
		return ""
	}

	return localPath
}
