package main

import (
	"strings"
	"testing"
)

func pointer[T any](value T) *T {
	return &value
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
		SessionID:   "12345678-abcd",
		SessionName: &sessionName,
		Model: &Model{
			DisplayName:         "GPT-5.6 Sol",
			ThinkingEffortLevel: pointer("high"),
		},
		Cost: &Cost{
			TotalDurationMS:      125_000,
			TotalLinesAdded:      12,
			TotalLinesRemoved:    3,
			TotalPremiumRequests: 2,
		},
		AIUsed: &AIUsed{Formatted: "1.25"},
		ContextWindow: &ContextWindow{
			CurrentContextUsedPercentage: pointer(42.0),
			TotalCacheReadTokens:         12_500,
			TotalCacheWriteTokens:        1_500,
			TotalReasoningTokens:         3_000,
		},
	}

	got := render(data, "feature/statusline")
	for _, want := range []string{
		"Add status line",
		"[GPT-5.6 Sol]",
		"#12345678",
		"\U0001F4C1 /tmp/project",
		"\U0001F33F feature/statusline",
		"42%",
		"1.25 AI",
		"2m5s",
		"\u26A1high",
		"\u219312.5k \u21911.5k",
		"\U0001F4AD 3.0k",
		"+12",
		"-3",
		"\u2B50 2",
	} {
		if !strings.Contains(got, want) {
			t.Errorf("rendered output missing %q:\n%s", want, got)
		}
	}
}

func TestRenderMinimalPayload(t *testing.T) {
	got := render(StatusData{}, "")
	if !strings.Contains(got, "[]") || !strings.Contains(got, "0%") {
		t.Fatalf("rendered minimal output is incomplete:\n%s", got)
	}
}
