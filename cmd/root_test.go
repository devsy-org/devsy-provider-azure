package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/devsy-org/devsy/logging"
	"github.com/devsy-org/devsy/pkg/log"
	"github.com/spf13/cobra"
)

const (
	testErrorLevel = "error"
	testJSONFormat = "json"
	testTextFormat = "text"
)

func TestProviderLogProcess(t *testing.T) {
	if os.Getenv("PROVIDER_LOG_TEST") != "1" {
		return
	}

	root := NewRootCmd()
	root.AddCommand(&cobra.Command{
		Use: "probe",
		RunE: func(_ *cobra.Command, _ []string) error {
			if _, err := fmt.Fprintln(os.Stdout, "command ran"); err != nil {
				return err
			}
			logger := log.Slog()
			for _, entry := range []struct {
				level   log.Level
				message string
			}{
				{logging.LevelTrace, "trace"},
				{slog.LevelDebug, "debug"},
				{slog.LevelInfo, "info"},
				{slog.LevelWarn, "warn"},
				{slog.LevelError, testErrorLevel},
			} {
				logger.Log(
					context.Background(),
					entry.level,
					entry.message+" "+os.Getenv("AZURE_CLIENT_SECRET"),
				)
			}
			_ = log.Sync()
			return nil
		},
	})
	root.SetArgs([]string{"probe"})
	if err := root.Execute(); err != nil {
		_, _ = fmt.Fprintln(os.Stdout, "error:", err)
	}
}

func TestProviderLogging(t *testing.T) {
	for _, tc := range []struct {
		level string
		want  string
	}{
		{"", "info,warn,error"},
		{testErrorLevel, testErrorLevel},
		{"warn", "warn,error"},
		{"info", "info,warn,error"},
		{"debug", "debug,info,warn,error"},
		{"trace", "trace,debug,info,warn,error"},
		{"TRACE", "trace,debug,info,warn,error"},
	} {
		t.Run("level="+tc.level, func(t *testing.T) {
			t.Setenv("PROVIDER_LOG_TEST", "1")
			t.Setenv("DEVSY_LOG_LEVEL", tc.level)
			t.Setenv("DEVSY_LOG_FORMAT", "auto")
			canary := t.Name() + "-private-value"
			t.Setenv("AZURE_CLIENT_SECRET", canary)
			command := providerLogCommand(t)
			var stdout, stderr bytes.Buffer
			command.Stdout, command.Stderr = &stdout, &stderr
			if err := command.Run(); err != nil {
				t.Fatal(err)
			}
			if !strings.Contains(stdout.String(), "command ran") {
				t.Fatalf("command did not run: %s", stdout.String())
			}
			if strings.Contains(stderr.String(), canary) {
				t.Fatal("secret leaked in diagnostics")
			}
			levels := providerLogLevels(t, stderr.String())
			if got := strings.Join(levels, ","); got != tc.want {
				t.Fatalf("levels %q, want %q", got, tc.want)
			}
		})
	}
}

func TestProviderLogConfig(t *testing.T) {
	for _, tc := range []struct {
		format string
		want   string
	}{
		{"", testJSONFormat},
		{"auto", testJSONFormat},
		{"JSON", testJSONFormat},
		{testTextFormat, testTextFormat},
		{"PLAIN", testTextFormat},
		{"logfmt", "logfmt"},
	} {
		t.Run(tc.format, func(t *testing.T) {
			t.Setenv("DEVSY_LOG_LEVEL", "")
			t.Setenv("DEVSY_LOG_FORMAT", tc.format)
			cfg, err := providerLogConfig()
			if err != nil || cfg.Format != tc.want {
				t.Fatalf("format %q: config %q, error %v", tc.format, cfg.Format, err)
			}
		})
	}
}

func TestInvalidProviderLoggingStopsCommand(t *testing.T) {
	for _, key := range []string{"DEVSY_LOG_LEVEL", "DEVSY_LOG_FORMAT"} {
		t.Run(key, func(t *testing.T) {
			t.Setenv("PROVIDER_LOG_TEST", "1")
			t.Setenv("DEVSY_LOG_LEVEL", "")
			t.Setenv("DEVSY_LOG_FORMAT", "")
			t.Setenv(key, "invalid-secret-canary")
			command := providerLogCommand(t)
			output, err := command.CombinedOutput()
			if err != nil {
				t.Fatal(err)
			}
			if strings.Contains(string(output), "command ran") ||
				strings.Contains(string(output), "invalid-secret-canary") ||
				!strings.Contains(string(output), "invalid "+key) {
				t.Fatalf("unexpected output: %s", output)
			}
		})
	}
}

func providerLogLevels(t *testing.T, output string) []string {
	t.Helper()
	var levels []string
	for line := range strings.SplitSeq(strings.TrimSpace(output), "\n") {
		var record struct {
			Time    string `json:"ts"`
			Level   string `json:"level"`
			Message string `json:"msg"`
		}
		if err := json.Unmarshal([]byte(line), &record); err != nil {
			t.Fatalf("invalid JSON diagnostic %q: %v", line, err)
		}
		if record.Time == "" || !strings.HasPrefix(record.Message, record.Level+" ") {
			t.Fatalf("invalid diagnostic: %s", line)
		}
		levels = append(levels, record.Level)
	}
	return levels
}

func providerLogCommand(t *testing.T) *exec.Cmd {
	t.Helper()
	executable, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	return &exec.Cmd{
		Path: executable,
		Args: []string{executable, "-test.run=^TestProviderLogProcess$"},
	}
}
