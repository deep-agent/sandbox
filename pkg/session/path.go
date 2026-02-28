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

func GetSessionWorkspace(rootWorkspace, sessionID string) string {
	if sessionID == "" {
		return rootWorkspace
	}
	return filepath.Join(rootWorkspace, sessionID)
}

func GetWorkspaceFromHeader(header http.Header, rootWorkspace string) string {
	return GetSessionWorkspace(rootWorkspace, GetSessionID(header))
}
