package main

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type Model struct {
	ID                  string  `json:"id"`
	DisplayName         string  `json:"display_name"`
	ThinkingEffort      *string `json:"thinking_effort"`
	ThinkingEffortLevel *string `json:"thinking_effort_level"`
	ReasoningEffort     *string `json:"reasoning_effort"`
}

type Cost struct {
	TotalDurationMS      int64 `json:"total_duration_ms"`
	TotalLinesAdded      int64 `json:"total_lines_added"`
	TotalLinesRemoved    int64 `json:"total_lines_removed"`
	TotalPremiumRequests int64 `json:"total_premium_requests"`
}

type AIUsed struct {
	Formatted    string   `json:"formatted"`
	TotalNanoAIU *float64 `json:"total_nano_aiu"`
}

type CurrentUsage struct {
	CacheCreationInputTokens int64 `json:"cache_creation_input_tokens"`
	CacheReadInputTokens     int64 `json:"cache_read_input_tokens"`
}

type ContextWindow struct {
	ContextWindowSize            *float64      `json:"context_window_size"`
	UsedPercentage               *float64      `json:"used_percentage"`
	TotalCacheReadTokens         int64         `json:"total_cache_read_tokens"`
	TotalCacheWriteTokens        int64         `json:"total_cache_write_tokens"`
	TotalReasoningTokens         int64         `json:"total_reasoning_tokens"`
	CurrentContextTokens         *float64      `json:"current_context_tokens"`
	CurrentContextUsedPercentage *float64      `json:"current_context_used_percentage"`
	DisplayedContextLimit        *float64      `json:"displayed_context_limit"`
	CurrentUsage                 *CurrentUsage `json:"current_usage"`
}

type StatusData struct {
	CWD         string  `json:"cwd"`
	SessionID   string  `json:"session_id"`
	SessionName *string `json:"session_name"`
	Model       *Model  `json:"model"`
	Workspace   *struct {
		CurrentDir string `json:"current_dir"`
	} `json:"workspace"`
	Remote *struct {
		Connected bool   `json:"connected"`
		Indicator string `json:"indicator"`
		TaskName  string `json:"task_name"`
	} `json:"remote"`
	Cost          *Cost          `json:"cost"`
	AIUsed        *AIUsed        `json:"ai_used"`
	ContextWindow *ContextWindow `json:"context_window"`
}

const (
	cyan   = "\033[36m"
	green  = "\033[32m"
	yellow = "\033[33m"
	red    = "\033[31m"
	reset  = "\033[0m"
)

func thermometer(pct float64, width int) string {
	pct = max(0, min(pct, 100))
	filled := int(math.Round(pct * float64(width) / 100))
	color := green
	if pct >= 90 {
		color = red
	} else if pct >= 70 {
		color = yellow
	}
	bar := strings.Repeat("\u2588", filled) + strings.Repeat("\u2591", width-filled)
	return fmt.Sprintf("%s%s%s", color, bar, reset)
}

func abbreviateHome(path string) string {
	if home, err := os.UserHomeDir(); err == nil && home != "" {
		if path == home {
			return "~"
		}
		if strings.HasPrefix(path, home+string(filepath.Separator)) {
			return "~" + path[len(home):]
		}
	}
	return path
}

func formatTokens(n int64) string {
	switch {
	case n >= 1_000_000:
		return fmt.Sprintf("%.1fM", float64(n)/1_000_000)
	case n >= 1_000:
		return fmt.Sprintf("%.1fk", float64(n)/1_000)
	default:
		return fmt.Sprintf("%d", n)
	}
}

func formatDuration(ms int64) string {
	totalSeconds := ms / 1000
	hours := totalSeconds / 3600
	minutes := (totalSeconds % 3600) / 60
	seconds := totalSeconds % 60
	if hours > 0 {
		return fmt.Sprintf("%dh%dm%ds", hours, minutes, seconds)
	}
	return fmt.Sprintf("%dm%ds", minutes, seconds)
}

func workingDir(data StatusData) string {
	if data.CWD != "" {
		return data.CWD
	}
	if data.Workspace != nil {
		return data.Workspace.CurrentDir
	}
	return ""
}

func contextPercentage(context *ContextWindow) float64 {
	if context == nil {
		return 0
	}
	if context.CurrentContextUsedPercentage != nil {
		return *context.CurrentContextUsedPercentage
	}
	if context.CurrentContextTokens != nil {
		limit := context.DisplayedContextLimit
		if limit == nil {
			limit = context.ContextWindowSize
		}
		if limit != nil && *limit > 0 {
			return *context.CurrentContextTokens * 100 / *limit
		}
	}
	if context.UsedPercentage != nil {
		return *context.UsedPercentage
	}
	return 0
}

func thinkingEffort(model *Model) string {
	if model == nil {
		return ""
	}
	for _, effort := range []*string{
		model.ThinkingEffortLevel,
		model.ThinkingEffort,
		model.ReasoningEffort,
	} {
		if effort != nil && *effort != "" {
			return *effort
		}
	}
	if _, effort, ok := strings.Cut(model.DisplayName, " \u00b7 "); ok {
		return strings.TrimSpace(effort)
	}
	return ""
}

func modelName(model *Model) string {
	if model == nil {
		return ""
	}

	name, _, _ := strings.Cut(model.DisplayName, " \u00b7 ")
	if name == "" {
		name = model.ID
	}
	if name != strings.ToLower(name) {
		return name
	}

	parts := strings.Split(strings.ToLower(name), "-")
	for i, part := range parts {
		switch part {
		case "gpt":
			parts[i] = "GPT"
		case "sol":
			parts[i] = "Sol"
		case "terra":
			parts[i] = "Terra"
		case "luna":
			parts[i] = "Luna"
		case "codex":
			parts[i] = "Codex"
		case "mini":
			parts[i] = "mini"
		default:
			if len(part) > 0 && part[0] >= 'a' && part[0] <= 'z' {
				parts[i] = strings.ToUpper(part[:1]) + part[1:]
			}
		}
	}

	if len(parts) > 1 && parts[0] == "GPT" {
		return parts[0] + "-" + parts[1] + " " + strings.Join(parts[2:], " ")
	}
	return strings.Join(parts, " ")
}

func render(data StatusData, branch string) string {
	model := modelName(data.Model)

	sessionPrefix := ""
	if data.SessionName != nil && *data.SessionName != "" {
		sessionPrefix = fmt.Sprintf("%s\U0001F3F7\uFE0F %s%s | ", cyan, *data.SessionName, reset)
	}

	sessionSuffix := ""
	if data.SessionID != "" {
		id := data.SessionID
		if len(id) > 8 {
			id = id[:8]
		}
		sessionSuffix = fmt.Sprintf(" %s#%s%s", cyan, id, reset)
	}

	remoteSuffix := ""
	if data.Remote != nil && data.Remote.Connected {
		remote := data.Remote.Indicator
		if remote == "" {
			remote = data.Remote.TaskName
		}
		if remote == "" {
			remote = "remote"
		}
		remoteSuffix = fmt.Sprintf(" | \u2601\uFE0F %s", remote)
	}

	branchSuffix := ""
	if branch != "" {
		branchSuffix = " | \U0001F33F " + branch
	}

	dir := abbreviateHome(workingDir(data))
	line1 := fmt.Sprintf("%s%s[%s]%s%s \U0001F4C1 %s%s%s",
		sessionPrefix, cyan, model, reset, sessionSuffix, dir, remoteSuffix, branchSuffix)

	pct := contextPercentage(data.ContextWindow)
	line2 := fmt.Sprintf("%s %d%%", thermometer(pct, 10), int(pct))

	if data.AIUsed != nil {
		credits := data.AIUsed.Formatted
		if credits == "" && data.AIUsed.TotalNanoAIU != nil {
			credits = fmt.Sprintf("%.2f", *data.AIUsed.TotalNanoAIU/1_000_000_000)
		}
		if credits != "" {
			line2 += fmt.Sprintf(" | %s%s AI%s", yellow, credits, reset)
		}
	}

	if data.Cost != nil {
		line2 += " | \u23F1\uFE0F " + formatDuration(data.Cost.TotalDurationMS)
	}

	if effort := thinkingEffort(data.Model); effort != "" {
		line2 += " | \u26A1" + effort
	}

	if context := data.ContextWindow; context != nil {
		cacheRead := context.TotalCacheReadTokens
		cacheWrite := context.TotalCacheWriteTokens
		if context.CurrentUsage != nil && cacheRead == 0 && cacheWrite == 0 {
			cacheRead = context.CurrentUsage.CacheReadInputTokens
			cacheWrite = context.CurrentUsage.CacheCreationInputTokens
		}
		if cacheRead != 0 || cacheWrite != 0 {
			line2 += fmt.Sprintf(" | \U0001F9E0 \u2193%s \u2191%s",
				formatTokens(cacheRead), formatTokens(cacheWrite))
		}
		if context.TotalReasoningTokens > 0 {
			line2 += " | \U0001F4AD " + formatTokens(context.TotalReasoningTokens)
		}
	}

	if data.Cost != nil {
		if data.Cost.TotalLinesAdded != 0 || data.Cost.TotalLinesRemoved != 0 {
			line2 += fmt.Sprintf(" | %s+%d%s %s-%d%s",
				green, data.Cost.TotalLinesAdded, reset,
				red, data.Cost.TotalLinesRemoved, reset)
		}
		if data.Cost.TotalPremiumRequests > 0 {
			line2 += fmt.Sprintf(" | \u2B50 %d", data.Cost.TotalPremiumRequests)
		}
	}

	return line1 + "\n" + line2
}

func gitBranch(dir string) string {
	if dir == "" {
		return ""
	}
	command := exec.Command("git", "branch", "--show-current")
	command.Dir = dir
	out, err := command.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}

func main() {
	var data StatusData
	if err := json.NewDecoder(os.Stdin).Decode(&data); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	fmt.Println(render(data, gitBranch(workingDir(data))))
}
