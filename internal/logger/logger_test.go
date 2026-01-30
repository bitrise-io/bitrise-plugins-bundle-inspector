package logger

import (
	"bytes"
	"strings"
	"testing"
)

func TestDefaultLogger(t *testing.T) {
	tests := []struct {
		name          string
		level         Level
		logFunc       func(Logger)
		wantContains  string
		wantEmpty     bool
	}{
		{
			name:  "info message at info level",
			level: LevelInfo,
			logFunc: func(l Logger) {
				l.Info("test info message")
			},
			wantContains: "test info message",
		},
		{
			name:  "warn message at info level",
			level: LevelInfo,
			logFunc: func(l Logger) {
				l.Warn("test warn message")
			},
			wantContains: "test warn message",
		},
		{
			name:  "error message at info level",
			level: LevelInfo,
			logFunc: func(l Logger) {
				l.Error("test error message")
			},
			wantContains: "test error message",
		},
		{
			name:  "debug message filtered at info level",
			level: LevelInfo,
			logFunc: func(l Logger) {
				l.Debug("test debug message")
			},
			wantEmpty: true,
		},
		{
			name:  "debug message at debug level",
			level: LevelDebug,
			logFunc: func(l Logger) {
				l.Debug("test debug message")
			},
			wantContains: "test debug message",
		},
		{
			name:  "info message filtered at error level",
			level: LevelError,
			logFunc: func(l Logger) {
				l.Info("test info message")
			},
			wantEmpty: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			logger := NewDefaultLogger(&buf, tt.level)

			tt.logFunc(logger)

			output := buf.String()
			if tt.wantEmpty {
				if output != "" {
					t.Errorf("Expected empty output, got %q", output)
				}
				return
			}

			if !strings.Contains(output, tt.wantContains) {
				t.Errorf("Output %q does not contain %q", output, tt.wantContains)
			}
		})
	}
}

func TestSilentLogger(t *testing.T) {
	logger := NewSilentLogger()

	// Should not panic
	logger.Debug("test")
	logger.Info("test")
	logger.Warn("test")
	logger.Error("test")
}

func TestGlobalLogger(t *testing.T) {
	// Save original logger
	original := global
	defer SetGlobal(original)

	var buf bytes.Buffer
	SetGlobal(NewDefaultLogger(&buf, LevelInfo))

	Info("test message")

	output := buf.String()
	if !strings.Contains(output, "test message") {
		t.Errorf("Global logger output %q does not contain 'test message'", output)
	}
}

func TestLoggerWithFormatting(t *testing.T) {
	var buf bytes.Buffer
	logger := NewDefaultLogger(&buf, LevelInfo)

	logger.Info("count: %d, name: %s", 42, "test")

	output := buf.String()
	if !strings.Contains(output, "count: 42") {
		t.Errorf("Output %q does not contain formatted number", output)
	}
	if !strings.Contains(output, "name: test") {
		t.Errorf("Output %q does not contain formatted string", output)
	}
}
