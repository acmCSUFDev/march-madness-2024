package badgerstub

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/dgraph-io/badger/v4"
)

// Logger is a badger.Logger implementation that uses slog.
// It treats INFO level as DEBUG.
type Logger struct {
	logger *slog.Logger
}

var _ badger.Logger = (*Logger)(nil)

func New(logger *slog.Logger) *Logger {
	return &Logger{logger: logger}
}

func (l *Logger) Errorf(format string, args ...interface{}) {
	l.logWithLevel(slog.LevelError, format, args...)
}

func (l *Logger) Warningf(format string, args ...interface{}) {
	l.logWithLevel(slog.LevelWarn, format, args...)
}

func (l *Logger) Infof(format string, args ...interface{}) {
	l.logWithLevel(slog.LevelDebug, format, args...)
}

func (l *Logger) Debugf(format string, args ...interface{}) {
	// l.logWithLevel(slog.LevelDebug, format, args...)
}

func (l *Logger) logWithLevel(level slog.Level, format string, args ...interface{}) {
	block := fmt.Sprintf(strings.Trim(format, "\n"), args...)
	for _, line := range strings.Split(block, "\n") {
		l.logger.Log(context.Background(), level, line)
	}
}
