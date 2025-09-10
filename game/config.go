package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

// GameConfig holds tunable settings loaded from env / defaults
type GameConfig struct {
	MaxTurns           int
	MetricMin          int
	MetricMax          int
	UseNarrativeEvents bool
	UseDirectorEvents  bool
}

func loadGameConfig() *GameConfig {
	cfg := &GameConfig{MaxTurns: 5, MetricMin: 40, MetricMax: 70, UseNarrativeEvents: true, UseDirectorEvents: true}
	if v := os.Getenv("PRES_SIM_MAX_TURNS"); v != "" { if i,err:=strconv.Atoi(v); err==nil && i>0 { cfg.MaxTurns = i } }
	if v := os.Getenv("PRES_SIM_METRIC_MIN"); v != "" { if i,err:=strconv.Atoi(v); err==nil { cfg.MetricMin = i } }
	if v := os.Getenv("PRES_SIM_METRIC_MAX"); v != "" { if i,err:=strconv.Atoi(v); err==nil { cfg.MetricMax = i } }
	if v := os.Getenv("PRES_SIM_USE_NARRATIVE"); v != "" { vv := strings.ToLower(v); cfg.UseNarrativeEvents = vv=="1" || vv=="true" || vv=="yes" }
	if v := os.Getenv("PRES_SIM_USE_DIRECTOR"); v != "" { vv := strings.ToLower(v); cfg.UseDirectorEvents = vv=="1" || vv=="true" || vv=="yes" }
	return cfg
}

// loadDotEnv loads key=value pairs from .env into environment
func loadDotEnv() {
	paths := []string{".env", "../.env", "../../.env", "game/.env"}
	if exe, err := os.Executable(); err == nil {
		dir := filepath.Dir(exe)
		paths = append(paths, filepath.Join(dir, ".env"))
		paths = append(paths, filepath.Join(dir, "../.env"))
	}
	loaded := false
	for _, p := range paths {
		f, err := os.Open(p)
		if err != nil { continue }
		s := bufio.NewScanner(f)
		for s.Scan() {
			line := strings.TrimSpace(s.Text())
			if line == "" || strings.HasPrefix(line, "#") { continue }
			parts := strings.SplitN(line, "=", 2)
			if len(parts) != 2 { continue }
			k := strings.TrimSpace(parts[0])
			v := strings.TrimSpace(parts[1])
			if k != "" { os.Setenv(k, v) }
		}
		f.Close()
		if !loaded { fmt.Printf("Loaded .env from %s\n", p); }
		loaded = true
		if os.Getenv("THETA_API_KEY") != "" { break }
	}
	if os.Getenv("THETA_API_KEY") == "" {
		fmt.Println("No THETA_API_KEY found after scanning paths:", paths)
	}
}
