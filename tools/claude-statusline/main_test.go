package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func pointer[T any](value T) *T {
	return &value
}

func decodeStatus(t *testing.T, payload string) StatusData {
	t.Helper()
	var data StatusData
	if err := json.Unmarshal([]byte(payload), &data); err != nil {
		t.Fatal(err)
	}
	return data
}

func TestRenderModelGroup(t *testing.T) {
	for _, effort := range []string{"low", "medium", "high", "xhigh", "max"} {
		t.Run(effort, func(t *testing.T) {
			data := decodeStatus(t, `{
				"model": {"id": "claude-opus-5-5", "display_name": "Opus"},
				"thinking": {"enabled": true},
				"fast_mode": true
			}`)
			data.Effort = &Effort{Level: effort}
			lines := strings.Split(render(data, ""), "\n")
			if len(lines) != 2 {
				t.Fatalf("want exactly two lines, got %d", len(lines))
			}
			want := "[Opus 5.5 \u26A1" + effort + " \U0001F4AD fast]"
			if !strings.Contains(lines[0], want) {
				t.Errorf("model group missing %q: %s", want, lines[0])
			}
			for _, marker := range []string{"\u26A1", "\U0001F4AD", "fast", "Opus"} {
				if strings.Contains(lines[1], marker) {
					t.Errorf("model detail %q leaked onto second line: %s", marker, lines[1])
				}
			}
		})
	}
}

func TestModelName(t *testing.T) {
	for _, tt := range []struct {
		name        string
		displayName string
		id          string
		want        string
	}{
		{"opus version", "Opus", "claude-opus-5-5", "Opus 5.5"},
		{"sonnet version", "Sonnet", "claude-sonnet-4-6", "Sonnet 4.6"},
		{"dated haiku", "Haiku", "claude-haiku-4-5-20251001", "Haiku 4.5"},
		{"major version", "Opus", "claude-opus-4-20250514", "Opus 4"},
		{"extended context", "Opus", "claude-opus-5-5[1m]", "Opus 5.5"},
		{"dotted version", "Opus", "claude-opus-5.5", "Opus 5.5"},
		{"latest suffix", "Opus", "claude-opus-5-5-latest", "Opus 5.5"},
		{"already versioned", "Opus 5.5", "claude-opus-5-5", "Opus 5.5"},
		{"custom display", "My Opus", "claude-opus-5-5", "My Opus"},
		{"different family", "Sonnet", "claude-opus-5-5", "Sonnet"},
		{"unversioned alias", "Opus", "claude-opus", "Opus"},
		{"unknown ID", "Opus", "custom-model", "Opus"},
		{"missing ID", "Opus", "", "Opus"},
		{"ID fallback", "", "claude-opus-5-5", "claude-opus-5-5"},
		{"missing model", "", "", ""},
	} {
		t.Run(tt.name, func(t *testing.T) {
			if got := modelName(tt.displayName, tt.id); got != tt.want {
				t.Errorf("modelName(%q, %q) = %q, want %q", tt.displayName, tt.id, got, tt.want)
			}
		})
	}
}

func TestRenderOptionalModelFields(t *testing.T) {
	for _, tt := range []struct {
		name    string
		payload string
		want    string
	}{
		{"minimal", `{}`, "[]"},
		{"display name", `{"model":{"display_name":"Sonnet"}}`, "[Sonnet]"},
		{"ID fallback", `{"model":{"id":"custom-model"}}`, "[custom-model]"},
		{"disabled modes", `{
			"model":{"display_name":"Sonnet"},
			"effort":{"level":""},
			"thinking":{"enabled":false},
			"fast_mode":false
		}`, "[Sonnet]"},
		{"null fields", `{
			"model":{"display_name":"Sonnet"},
			"effort":null,
			"thinking":null,
			"context_window":null
		}`, "[Sonnet]"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			got := render(decodeStatus(t, tt.payload), "")
			if !strings.Contains(got, tt.want) {
				t.Errorf("render missing %q:\n%s", tt.want, got)
			}
			if !strings.Contains(got, "0%") {
				t.Errorf("missing initial context percentage:\n%s", got)
			}
		})
	}
}

func TestContextDetail(t *testing.T) {
	for _, tt := range []struct {
		name    string
		payload string
		want    string
	}{
		{"input and caches only", `{
			"context_window_size":200000, "used_percentage":42,
			"total_input_tokens":84000, "total_output_tokens":5000,
			"current_usage":{"input_tokens":4000,"output_tokens":5000,
				"cache_creation_input_tokens":10000,"cache_read_input_tokens":70000}
		}`, " (84.0k/200.0k)"},
		{"extended context totals", `{
			"context_window_size":1000000,"used_percentage":25,
			"total_input_tokens":250000,"total_output_tokens":50000
		}`, " (250.0k/1.0M)"},
		{"zero usage", `{
			"context_window_size":200000,"used_percentage":0,
			"total_input_tokens":0
		}`, " (0/200.0k)"},
		{"startup", `{
			"context_window_size":1000000,"used_percentage":null,
			"current_usage":null,"total_input_tokens":0
		}`, " (1.0M ctx)"},
		{"after compact", `{
			"context_window_size":200000,"used_percentage":null,
			"current_usage":null,"total_input_tokens":84000
		}`, " (200.0k ctx)"},
		{"no token counts", `{
			"context_window_size":200000,"used_percentage":42
		}`, " (200.0k ctx)"},
		{"missing capacity", `{"used_percentage":42}`, ""},
		{"nonpositive capacity", `{"context_window_size":-1}`, ""},
	} {
		t.Run(tt.name, func(t *testing.T) {
			data := decodeStatus(t, `{"context_window":`+tt.payload+`}`)
			if got := contextDetail(data.ContextWindow); got != tt.want {
				t.Errorf("contextDetail() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestRenderPreservesSessionAndMetrics(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	data := decodeStatus(t, `{
		"model":{"display_name":"Opus"},
		"session_id":"12345678-abcd","session_name":"Update statusline",
		"workspace":{"current_dir":"/tmp/project/subdir","project_dir":"/tmp/project",
			"added_dirs":["/tmp/other","/tmp/extra"]},
		"cost":{"total_cost_usd":1.25,"total_duration_ms":125000},
		"context_window":{"context_window_size":1000000,"used_percentage":42,
			"current_usage":{"input_tokens":400000,
				"cache_read_input_tokens":12500,"cache_creation_input_tokens":7500}},
		"exceeds_200k_tokens":true,
		"rate_limits":{"five_hour":{"used_percentage":23.5},
			"seven_day":{"used_percentage":41.2}}
	}`)

	got := render(data, "feature/statusline")
	for _, want := range []string{
		"Update statusline", "#12345678", "\U0001F4C1 /tmp/project/subdir",
		"[project]", "+[other,extra]", "\U0001F33F feature/statusline",
		"42% (420.0k/1.0M)", "$1.25", "2m5s",
		"\u219312.5k \u21917.5k", "\u26A0\uFE0F200k+", "5h:", "7d:",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("render missing %q:\n%s", want, got)
		}
	}
}

func TestRenderPeerSessionName(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	sessions := filepath.Join(home, ".claude", "sessions")
	if err := os.MkdirAll(sessions, 0700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(sessions, "peer.json"),
		[]byte(`{"sessionId":"12345678-abcd","name":"project-cf"}`), 0600); err != nil {
		t.Fatal(err)
	}
	got := render(StatusData{SessionID: pointer("12345678-abcd")}, "")
	if !strings.Contains(got, "@project-cf") || strings.Contains(got, "#12345678") {
		t.Errorf("peer session name was not preserved:\n%s", got)
	}
}

func TestFormatTokens(t *testing.T) {
	for _, tt := range []struct {
		tokens int64
		want   string
	}{
		{0, "0"}, {999, "999"}, {1000, "1.0k"}, {12500, "12.5k"},
		{200000, "200.0k"}, {1000000, "1.0M"}, {1500000, "1.5M"},
	} {
		if got := formatTokens(tt.tokens); got != tt.want {
			t.Errorf("formatTokens(%d) = %q, want %q", tt.tokens, got, tt.want)
		}
	}
}
