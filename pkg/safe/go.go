package safe

import (
	"runtime/debug"

	"github.com/deep-agent/sandbox/pkg/logger"
)

func Go(fn func()) {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				logger.Printf("[PANIC] goroutine panic: %v\n%s", r, debug.Stack())
			}
		}()
		fn()
	}()
}

func GoWithRecover(fn func(), onPanic func(r interface{})) {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				if onPanic != nil {
					onPanic(r)
				} else {
					logger.Printf("[PANIC] goroutine panic: %v\n%s", r, debug.Stack())
				}
			}
		}()
		fn()
	}()
}
