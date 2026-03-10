package logger

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"strings"
	"time"
)

// defaultLogger should only be changed via SetDefault during main init, before spawning goroutines.
var defaultLogger *slog.Logger

func init() {
	defaultLogger = slog.New(&plainHandler{w: os.Stdout, level: slog.LevelInfo})
}

// plainHandler outputs logs in a simple plain-text format that preserves newlines in messages.
type plainHandler struct {
	w     io.Writer
	level slog.Level
}

func (h *plainHandler) Enabled(_ context.Context, level slog.Level) bool {
	return level >= h.level
}

func (h *plainHandler) Handle(_ context.Context, r slog.Record) error {
	ts := r.Time.Format(time.DateTime)
	var b strings.Builder
	fmt.Fprintf(&b, "%s %s %s", ts, r.Level.String(), r.Message)
	r.Attrs(func(a slog.Attr) bool {
		fmt.Fprintf(&b, " %s=%v", a.Key, a.Value)
		return true
	})
	b.WriteByte('\n')
	_, err := io.WriteString(h.w, b.String())
	return err
}

func (h *plainHandler) WithAttrs(_ []slog.Attr) slog.Handler { return h }
func (h *plainHandler) WithGroup(_ string) slog.Handler      { return h }

func SetDefault(l *slog.Logger) {
	defaultLogger = l
}

func Default() *slog.Logger {
	return defaultLogger
}

func Debug(msg string, args ...any) {
	defaultLogger.Debug(msg, args...)
}

func Info(msg string, args ...any) {
	defaultLogger.Info(msg, args...)
}

func Warn(msg string, args ...any) {
	defaultLogger.Warn(msg, args...)
}

func Error(msg string, args ...any) {
	defaultLogger.Error(msg, args...)
}

func Fatal(msg string, args ...any) {
	defaultLogger.Error(msg, args...)
	os.Exit(1)
}

func Print(args ...any) {
	defaultLogger.Info(fmt.Sprint(args...))
}

func Println(args ...any) {
	defaultLogger.Info(trimTrailingNewline(fmt.Sprintln(args...)))
}

func Printf(format string, args ...any) {
	defaultLogger.Info(fmt.Sprintf(format, args...))
}

func Panic(args ...any) {
	msg := fmt.Sprint(args...)
	defaultLogger.Error(msg)
	panic(msg)
}

func Panicln(args ...any) {
	msg := trimTrailingNewline(fmt.Sprintln(args...))
	defaultLogger.Error(msg)
	panic(msg)
}

func Panicf(format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	defaultLogger.Error(msg)
	panic(msg)
}

func Fatalf(format string, args ...any) {
	defaultLogger.Error(fmt.Sprintf(format, args...))
	os.Exit(1)
}

func Debugf(format string, args ...any) {
	defaultLogger.Debug(fmt.Sprintf(format, args...))
}

func Infof(format string, args ...any) {
	defaultLogger.Info(fmt.Sprintf(format, args...))
}

func Warnf(format string, args ...any) {
	defaultLogger.Warn(fmt.Sprintf(format, args...))
}

func Errorf(format string, args ...any) {
	defaultLogger.Error(fmt.Sprintf(format, args...))
}

func DebugfContext(ctx context.Context, format string, args ...any) {
	defaultLogger.DebugContext(ctx, fmt.Sprintf(format, args...))
}

func InfofContext(ctx context.Context, format string, args ...any) {
	defaultLogger.InfoContext(ctx, fmt.Sprintf(format, args...))
}

func WarnfContext(ctx context.Context, format string, args ...any) {
	defaultLogger.WarnContext(ctx, fmt.Sprintf(format, args...))
}

func ErrorfContext(ctx context.Context, format string, args ...any) {
	defaultLogger.ErrorContext(ctx, fmt.Sprintf(format, args...))
}

func trimTrailingNewline(s string) string {
	return strings.TrimSuffix(s, "\n")
}
