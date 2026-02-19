package utils

import (
	"testing"
)

func TestIsAudioFile(t *testing.T) {
	tests := []struct {
		filename    string
		contentType string
		want        bool
	}{
		{"song.mp3", "", true},
		{"voice.wav", "", true},
		{"audio.ogg", "", true},
		{"file.m4a", "", true},
		{"track.MP3", "", true},
		{"doc.pdf", "", false},
		{"image.png", "", false},
		{"", "audio/mpeg", true},
		{"", "audio/wav", true},
		{"", "application/ogg", true},
		{"", "application/x-ogg", true},
		{"", "AUDIO/MPEG", true},
		{"", "video/mp4", false},
	}
	for _, tt := range tests {
		got := IsAudioFile(tt.filename, tt.contentType)
		if got != tt.want {
			t.Errorf("IsAudioFile(%q, %q) = %v, want %v", tt.filename, tt.contentType, got, tt.want)
		}
	}
}

func TestSanitizeFilename(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"normal.txt", "normal.txt"},
		{"/path/to/file.txt", "file.txt"},
		{"..danger..", "danger"},
		{"a/b/c", "c"},
		{"a\\b\\c", "c"},
	}
	for _, tt := range tests {
		got := SanitizeFilename(tt.input)
		if got != tt.want {
			t.Errorf("SanitizeFilename(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}
