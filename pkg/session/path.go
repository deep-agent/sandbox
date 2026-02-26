package session

import (
	"net/http"
	"path/filepath"
)

const HeaderSessionID = "X-Session-ID"

type HertzHeader interface {
	Peek(key string) []byte
}

func GetSessionID(header http.Header) string {
	return header.Get(HeaderSessionID)
}

func GetSessionIDFromHertz(header HertzHeader) string {
	return string(header.Peek(HeaderSessionID))
}

func GetWorkspace(homeDir, sessionID string) string {
	if sessionID == "" {
		return homeDir
	}
	return filepath.Join(homeDir, sessionID)
}

func GetWorkspaceFromHeader(header http.Header, homeDir string) string {
	return GetWorkspace(homeDir, GetSessionID(header))
}
