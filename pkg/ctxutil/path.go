package ctxutil

import (
	"context"
	"fmt"
	"os"

	"github.com/deep-agent/sandbox/types/consts"
)

type cwdKey struct{}
type sessionIDKey struct{}

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

	envWorkspace := os.Getenv(consts.Workspace)
	sessionID := GetSessionIDFromCtx(ctx)
	if sessionID != "" {
		return fmt.Sprintf("%s/%s", envWorkspace, GetSessionIDFromCtx(ctx))
	}

	return os.Getenv(consts.Workspace)
}

func WithSessionID(ctx context.Context, sessionID string) context.Context {
	return context.WithValue(ctx, sessionIDKey{}, sessionID)
}

func GetSessionIDFromCtx(ctx context.Context) string {
	if sessionID, ok := ctx.Value(sessionIDKey{}).(string); ok {
		return sessionID
	}
	return ""
}
