package consts

import "time"

const (
	DefaultBashTimeout        = 30 * time.Second
	DefaultBashBackgroundTime = 10 * time.Minute
	MaxBashTimeout            = 60 * time.Second

	DefaultReadLimit   = 2000
	DefaultMaxLineLen  = 2000
	DefaultOutputLimit = 30000

	DefaultGrepLimit      = 100
	DefaultGrepMaxLineLen = 2000
	DefaultGlobLimit      = 100

	DefaultBrowserTimeout = 30 * time.Second
)
