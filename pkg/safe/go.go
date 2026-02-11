package safe

import (
	"log"
	"runtime/debug"
)

func Go(fn func()) {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("[PANIC] goroutine panic: %v\n%s", r, debug.Stack())
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
					log.Printf("[PANIC] goroutine panic: %v\n%s", r, debug.Stack())
				}
			}
		}()
		fn()
	}()
}
