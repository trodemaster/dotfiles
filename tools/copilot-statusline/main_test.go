package main

import (
	"encoding/json"
	"strings"
	"testing"
)

func pointer[T any](value T) *T {
	return &value
}

func plainOutput(output string) string {
	return strings.NewReplacer(cyan, "", green, "", yellow, "", red, "", reset, "").Replace(output)
}

func TestContextPercentagePrefersLiveUsage(t *testing.T) {
	context := &ContextWindow{
		ContextWindowSize:            pointer(400_000.0),
		UsedPercentage:               pointer(80.0),
		CurrentContextTokens:         pointer(25_000.0),
		DisplayedContextLimit:        pointer(1_050_000.0),
		CurrentContextUsedPercentage: pointer(2.38),
	}

	if got := contextPercentage(context); got != 2.38 {
		t.Fatalf("contextPercentage() = %v, want 2.38", got)
	}
}

func TestContextPercentageDerivesLiveUsage(t *testing.T) {
	context := &ContextWindow{
		CurrentContextTokens:  pointer(100_000.0),
		DisplayedContextLimit: pointer(400_000.0),
		UsedPercentage:        pointer(90.0),
	}

	if got := contextPercentage(context); got != 25 {
		t.Fatalf("contextPercentage() = %v, want 25", got)
	}
}

func TestModelNameAndEffortFromCopilotDisplayName(t *testing.T) {
	model := &Model{
		ID:          "gpt-5.6-sol",
		DisplayName: "gpt-5.6-sol \u00b7 medium",
	}

	if got := modelName(model); got != "GPT-5.6 Sol" {
		t.Fatalf("modelName() = %q, want %q", got, "GPT-5.6 Sol")
	}
	if got := thinkingEffort(model); got != "medium" {
		t.Fatalf("thinkingEffort() = %q, want %q", got, "medium")
	}
}

func TestRenderFullPayload(t *testing.T) {
	sessionName := "Add status line"
	data := StatusData{
		CWD:         "/tmp/project",
		SessionName: &sessionName,
		Model: &Model{
			DisplayName:         "GPT-5.6 Sol",
			ThinkingEffortLevel: pointer("high"),
		},
		Cost: &Cost{
			TotalPremiumRequests: 2,
		},
		AIUsed: &AIUsed{Formatted: "1.25"},
		ContextWindow: &ContextWindow{
			ContextWindowSize:            pointer(400_000.0),
			DisplayedContextLimit:        pointer(1_050_000.0),
			CurrentContextUsedPercentage: pointer(42.0),
			TotalCacheReadTokens:         12_500,
			TotalCacheWriteTokens:        1_500,
			TotalReasoningTokens:         3_000,
		},
	}

	got := plainOutput(render(data, "feature/statusline"))
	for _, want := range []string{
		"Add status line",
		"GPT-5.6 Sol \u2022 \u26A1high \u2022 1.1M",
		"\U0001F4C1 /tmp/project",
		"\U0001F33F feature/statusline",
		"42%",
		"1.25 \U0001F4B0",
		"\u219312.5k \u21911.5k",
		"\U0001F4AD 3.0k",
		"\U0001F4B5 2",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("rendered output missing %q:\n%s", want, got)
		}
	}
	lines := strings.Split(got, "\n")
	if len(lines) != 2 {
		t.Fatalf("want exactly two lines, got %d", len(lines))
	}
	if !strings.Contains(lines[0], "GPT-5.6 Sol \u2022 \u26A1high \u2022 1.1M") {
		t.Errorf("model segment is not on line 1: %s", lines[0])
	}
	if !strings.HasPrefix(lines[1], "\u2588\u2588\u2588\u2588\u2591\u2591\u2591\u2591\u2591\u2591 42%") {
		t.Errorf("context bar changed or moved: %s", lines[1])
	}
	for _, unwanted := range []string{"GPT-5.6 Sol", "\u26A1high", "1.1M"} {
		if strings.Contains(lines[1], unwanted) {
			t.Errorf("model detail %q still appears on line 2: %s", unwanted, lines[1])
		}
	}
	wantSuffix := " | \U0001F9E0 \u219312.5k \u21911.5k | \U0001F4AD 3.0k | 1.25 \U0001F4B0 | \U0001F4B5 2"
	if !strings.HasSuffix(lines[1], wantSuffix) {
		t.Errorf("cache and reasoning must precede the session usage and premium suffix: %s", lines[1])
	}
	for _, unwanted := range []string{" AI", "\u2B50", "\U0001F4B2"} {
		if strings.Contains(got, unwanted) {
			t.Errorf("old usage label %q still appears:\n%s", unwanted, got)
		}
	}
}

func TestRenderSessionUsageIcons(t *testing.T) {
	for _, tt := range []struct {
		name  string
		usage *AIUsed
		want  string
	}{
		{"formatted", &AIUsed{Formatted: "1.25", TotalNanoAIU: pointer(2_500_000_000.0)}, "1.25 \U0001F4B0"},
		{"nano units", &AIUsed{TotalNanoAIU: pointer(2_500_000_000.0)}, "2.50 \U0001F4B0"},
		{"zero usage", &AIUsed{TotalNanoAIU: pointer(0.0)}, "0.00 \U0001F4B0"},
		{"missing usage", nil, ""},
		{"empty usage", &AIUsed{}, ""},
	} {
		t.Run(tt.name, func(t *testing.T) {
			got := plainOutput(render(StatusData{
				AIUsed: tt.usage,
				Cost:   &Cost{TotalPremiumRequests: 2},
				ContextWindow: &ContextWindow{
					TotalReasoningTokens: 3_000,
				},
			}, ""))
			if tt.want == "" {
				if strings.Contains(got, "\U0001F4B0") {
					t.Errorf("unavailable session usage shown:\n%s", got)
				}
			} else if !strings.HasSuffix(got, " | \U0001F4AD 3.0k | "+tt.want+" | \U0001F4B5 2") {
				t.Errorf("session usage and premium icons missing or incorrect:\n%s", got)
			}
			if !strings.HasSuffix(got, " | \U0001F4B5 2") {
				t.Errorf("green cash premium icon missing:\n%s", got)
			}
		})
	}
}

func TestRenderSessionUsageWithoutPremiumRequests(t *testing.T) {
	for _, cost := range []*Cost{nil, {TotalPremiumRequests: 0}} {
		got := plainOutput(render(StatusData{
			AIUsed: &AIUsed{Formatted: "1.25"},
			Cost:   cost,
			ContextWindow: &ContextWindow{
				TotalReasoningTokens: 3_000,
			},
		}, ""))
		if !strings.HasSuffix(got, " | \U0001F4AD 3.0k | 1.25 \U0001F4B0") {
			t.Errorf("session usage is not at the end without premium requests:\n%s", got)
		}
		if strings.Contains(got, "\U0001F4B5") {
			t.Errorf("unavailable premium requests shown:\n%s", got)
		}
	}
}

func TestRenderCacheFallbackBeforePremiumRequests(t *testing.T) {
	got := plainOutput(render(StatusData{
		ContextWindow: &ContextWindow{
			CurrentUsage: &CurrentUsage{
				CacheReadInputTokens:     12_500,
				CacheCreationInputTokens: 1_500,
			},
			TotalReasoningTokens: 3_000,
		},
		Cost: &Cost{TotalPremiumRequests: 2},
	}, ""))
	wantSuffix := " | \U0001F9E0 \u219312.5k \u21911.5k | \U0001F4AD 3.0k | \U0001F4B5 2"
	if !strings.HasSuffix(got, wantSuffix) {
		t.Errorf("current-usage cache counts, reasoning, and premium requests missing or out of order:\n%s", got)
	}
}

func TestRenderMinimalPayload(t *testing.T) {
	got := plainOutput(render(StatusData{}, ""))
	want := "\U0001F4C1 \n" + strings.Repeat("\u2591", 10) + " 0%"
	if got != want {
		t.Fatalf("rendered minimal output is incomplete:\n%s", got)
	}
}

func TestModelSegmentOptionalFields(t *testing.T) {
	for _, tt := range []struct {
		name string
		data StatusData
		want string
	}{
		{"model only", StatusData{Model: &Model{DisplayName: "Opus 5.5"}}, "Opus 5.5"},
		{"effort from display name", StatusData{
			Model: &Model{DisplayName: "gpt-5.6-sol \u00b7 medium"},
		}, "GPT-5.6 Sol \u2022 \u26A1medium"},
		{"capacity fallback", StatusData{
			Model:         &Model{DisplayName: "Opus 5.5"},
			ContextWindow: &ContextWindow{ContextWindowSize: pointer(200_000.0)},
		}, "Opus 5.5 \u2022 200.0k"},
		{"unavailable displayed capacity", StatusData{
			Model: &Model{DisplayName: "Opus 5.5"},
			ContextWindow: &ContextWindow{
				DisplayedContextLimit: pointer(0.0),
				ContextWindowSize:     pointer(200_000.0),
			},
		}, "Opus 5.5"},
		{"capacity without model", StatusData{
			ContextWindow: &ContextWindow{DisplayedContextLimit: pointer(1_050_000.0)},
		}, "1.1M"},
		{"empty", StatusData{}, ""},
	} {
		t.Run(tt.name, func(t *testing.T) {
			if got := plainOutput(modelSegment(tt.data)); got != tt.want {
				t.Errorf("modelSegment() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestRenderOmitsRemovedFields(t *testing.T) {
	var data StatusData
	if err := json.Unmarshal([]byte(`{
		"model":{"display_name":"Claude Opus 5.5","thinking_effort":"high"},
		"session_id":"2bd6eb21-abcd",
		"context_window":{"displayed_context_limit":1050000,"current_context_used_percentage":25},
		"cost":{"total_duration_ms":3430000,"total_lines_added":820,
			"total_lines_removed":105,"total_premium_requests":75}
	}`), &data); err != nil {
		t.Fatal(err)
	}
	got := plainOutput(render(data, ""))
	if !strings.Contains(got, "Claude Opus 5.5 \u2022 \u26A1high \u2022 1.1M") {
		t.Errorf("compact model segment missing:\n%s", got)
	}
	for _, unwanted := range []string{"[", "]", "#2bd6eb21", "\u23F1", "57m10s", "+820", "-105", "context"} {
		if strings.Contains(got, unwanted) {
			t.Errorf("removed detail %q still appears:\n%s", unwanted, got)
		}
	}
}

func TestRenderRemoteIndicator(t *testing.T) {
	for _, tt := range []struct {
		name    string
		payload string
		want    string
	}{
		{"connected", `{"connected":true,"indicator":"remote-host","task_name":"task"}`, "remote-host"},
		{"task fallback", `{"connected":true,"task_name":"task"}`, "task"},
		{"generic fallback", `{"connected":true}`, "remote"},
		{"disabled", `{"connected":false,"indicator":"remote-host"}`, ""},
	} {
		t.Run(tt.name, func(t *testing.T) {
			var data StatusData
			if err := json.Unmarshal([]byte(`{"remote":`+tt.payload+`}`), &data); err != nil {
				t.Fatal(err)
			}
			got := render(data, "")
			if tt.want == "" {
				if strings.Contains(got, "\u2601") {
					t.Errorf("disconnected remote indicator shown: %s", got)
				}
			} else if !strings.Contains(got, "\u2601\uFE0F "+tt.want) {
				t.Errorf("remote indicator %q missing: %s", tt.want, got)
			}
		})
	}
}

func TestRenderContextBarColors(t *testing.T) {
	for _, tt := range []struct {
		name  string
		pct   float64
		color string
	}{
		{"low", 25, green},
		{"warning", 70, yellow},
		{"critical", 90, red},
	} {
		t.Run(tt.name, func(t *testing.T) {
			got := render(StatusData{ContextWindow: &ContextWindow{
				CurrentContextUsedPercentage: pointer(tt.pct),
			}}, "")
			lines := strings.Split(got, "\n")
			if !strings.HasPrefix(lines[1], tt.color) {
				t.Errorf("context bar has wrong color: %q", lines[1])
			}
		})
	}
}
