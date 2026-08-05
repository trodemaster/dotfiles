package main

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

type RateWindow struct {
	UsedPercentage *float64 `json:"used_percentage"`
	ResetsAt       *int64   `json:"resets_at"`
}

type CurrentUsage struct {
	CacheCreationInputTokens int64 `json:"cache_creation_input_tokens"`
	CacheReadInputTokens     int64 `json:"cache_read_input_tokens"`
}

type Effort struct {
	Level string `json:"level"`
}

type Thinking struct {
	Enabled bool `json:"enabled"`
}

type StatusData struct {
	Model struct {
		DisplayName string `json:"display_name"`
	} `json:"model"`
	Workspace struct {
		CurrentDir string   `json:"current_dir"`
		ProjectDir string   `json:"project_dir"`
		AddedDirs  []string `json:"added_dirs"`
	} `json:"workspace"`
	Cost struct {
		TotalCostUSD    float64 `json:"total_cost_usd"`
		TotalDurationMS int64   `json:"total_duration_ms"`
	} `json:"cost"`
	ContextWindow struct {
		UsedPercentage *float64      `json:"used_percentage"`
		CurrentUsage   *CurrentUsage `json:"current_usage"`
	} `json:"context_window"`
	RateLimits *struct {
		FiveHour *RateWindow `json:"five_hour"`
		SevenDay *RateWindow `json:"seven_day"`
	} `json:"rate_limits"`
	SessionName       *string   `json:"session_name"`
	Effort            *Effort   `json:"effort"`
	Thinking          *Thinking `json:"thinking"`
	Exceeds200kTokens bool      `json:"exceeds_200k_tokens"`
}

const (
	cyan   = "\033[36m"
	green  = "\033[32m"
	yellow = "\033[33m"
	red    = "\033[31m"
	reset  = "\033[0m"
)

func thermometer(pct float64, width int) string {
	filled := int(math.Round(pct * float64(width) / 100))
	if filled > width {
		filled = width
	}
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
	if n >= 1000 {
		return fmt.Sprintf("%.1fk", float64(n)/1000)
	}
	return fmt.Sprintf("%d", n)
}

func timeUntil(unixSecs int64) string {
	d := time.Until(time.Unix(unixSecs, 0))
	if d <= 0 {
		return "now"
	}
	totalMins := int(d.Minutes())
	hours := totalMins / 60
	mins := totalMins % 60
	days := hours / 24
	hours = hours % 24
	switch {
	case days > 0:
		return fmt.Sprintf("%dd%dh", days, hours)
	case hours > 0:
		return fmt.Sprintf("%dh%dm", hours, mins)
	default:
		return fmt.Sprintf("%dm", mins)
	}
}

func main() {
	var data StatusData
	if err := json.NewDecoder(os.Stdin).Decode(&data); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	model := data.Model.DisplayName
	dir := abbreviateHome(data.Workspace.CurrentDir)

	ctxPct := 0.0
	if data.ContextWindow.UsedPercentage != nil {
		ctxPct = *data.ContextWindow.UsedPercentage
	}

	cost := data.Cost.TotalCostUSD
	durationMS := data.Cost.TotalDurationMS
	mins := durationMS / 60000
	secs := (durationMS % 60000) / 1000

	// Git branch
	branch := ""
	if out, err := exec.Command("git", "branch", "--show-current").Output(); err == nil {
		if b := strings.TrimSpace(string(out)); b != "" {
			branch = " | \U0001F33F " + b
		}
	}

	// Session name prefix
	sessionPrefix := ""
	if data.SessionName != nil && *data.SessionName != "" {
		sessionPrefix = fmt.Sprintf("%s\U0001F3F7\uFE0F %s%s | ", cyan, *data.SessionName, reset)
	}

	// Project dir suffix (only when it differs from cwd)
	projectSuffix := ""
	if data.Workspace.ProjectDir != "" {
		if data.Workspace.ProjectDir != data.Workspace.CurrentDir {
			projectSuffix = fmt.Sprintf(" %s[%s]%s", yellow, filepath.Base(data.Workspace.ProjectDir), reset)
		}
	}

	// Added dirs suffix
	addedDirsSuffix := ""
	if len(data.Workspace.AddedDirs) > 0 {
		names := make([]string, len(data.Workspace.AddedDirs))
		for i, d := range data.Workspace.AddedDirs {
			names[i] = filepath.Base(d)
		}
		addedDirsSuffix = fmt.Sprintf(" +[%s]", strings.Join(names, ","))
	}

	// Line 1: session, model, dir, project, added dirs, git branch
	fmt.Printf("%s%s[%s]%s \U0001F4C1 %s%s%s%s\n",
		sessionPrefix, cyan, model, reset, dir, projectSuffix, addedDirsSuffix, branch)

	// Line 2: context bar | cost | duration | effort | thinking | 200k+ | rate limits
	ctxBar := thermometer(ctxPct, 10)
	line2 := fmt.Sprintf("%s %d%% | %s$%.2f%s | \u23F1\uFE0F %dm%ds",
		ctxBar, int(ctxPct), yellow, cost, reset, mins, secs)

	if data.Effort != nil && data.Effort.Level != "" {
		line2 += fmt.Sprintf(" | \u26A1%s", data.Effort.Level)
	}

	if data.Thinking != nil && data.Thinking.Enabled {
		line2 += " | \U0001F4AD"
	}

	if data.ContextWindow.CurrentUsage != nil {
		cu := data.ContextWindow.CurrentUsage
		line2 += fmt.Sprintf(" | \U0001F9E0 \u2193%s \u2191%s",
			formatTokens(cu.CacheReadInputTokens),
			formatTokens(cu.CacheCreationInputTokens))
	}

	if data.Exceeds200kTokens {
		line2 += fmt.Sprintf(" | %s\u26A0\uFE0F200k+%s", red, reset)
	}

	if data.RateLimits != nil {
		if w := data.RateLimits.FiveHour; w != nil && w.UsedPercentage != nil {
			bar := thermometer(*w.UsedPercentage, 5)
			reset_str := ""
			if w.ResetsAt != nil {
				reset_str = " " + timeUntil(*w.ResetsAt)
			}
			line2 += fmt.Sprintf(" | 5h:%s%d%%%s%s", bar, int(*w.UsedPercentage), reset, reset_str)
		}
		if w := data.RateLimits.SevenDay; w != nil && w.UsedPercentage != nil {
			bar := thermometer(*w.UsedPercentage, 5)
			reset_str := ""
			if w.ResetsAt != nil {
				reset_str = " " + timeUntil(*w.ResetsAt)
			}
			line2 += fmt.Sprintf(" | 7d:%s%d%%%s%s", bar, int(*w.UsedPercentage), reset, reset_str)
		}
	}

	fmt.Println(line2)
}
