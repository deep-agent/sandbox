package ctxutil

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"os"
)

type cwdKey struct{}
type logIDKey struct{}

func WithCwd(ctx context.Context, cwd string) context.Context {
	if cwd == "" {
		return ctx
	}
	return context.WithValue(ctx, cwdKey{}, cwd)
}

func GetCwd(ctx context.Context) string {
	if cwd, ok := ctx.Value(cwdKey{}).(string); ok {
		return cwd
	}
	return HomeDir()
}

func HomeDir() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return "/home"
	}
	return home
}

func WithLogID(ctx context.Context, logID string) context.Context {
	if logID == "" {
		return ctx
	}
	return context.WithValue(ctx, logIDKey{}, logID)
}

func GetLogIDFromCtx(ctx context.Context) string {
	if logID, ok := ctx.Value(logIDKey{}).(string); ok {
		return logID
	}
	return ""
}

func NewLogID() string {
	// 16 bytes => 32 hex chars
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		// Extremely unlikely; fall back to a deterministic non-empty value.
		return "00000000000000000000000000000000"
	}
	return hex.EncodeToString(b)
}
