package main

import (
	"fmt"
	"github.com/linlexing/dbgo/log"
	"runtime/debug"
)

// An error description, used as an argument to the error template.
type Error struct {
	err   interface{}
	stack []byte // The raw stack trace string from debug.Stack().
}

func NewError(stack []byte, err interface{}) *Error {
	return &Error{err: err, stack: stack}
}
func (e *Error) Error() string {
	return fmt.Sprintf("%v", e.err)
}

// PanicFilter wraps the action invocation in a protective defer blanket that
// converts panics into 500 error pages.
func PanicFilter(c *ControllerAgent, fc []Filter) {
	defer func() {
		if err := recover(); err != nil {
			handleInvocationPanic(c, err)
		}
	}()
	if c.Result == nil && len(fc) > 0 {
		fc[0](c, fc[1:])
	}
}

// This function handles a panic in an action invocation.
// It cleans up the stack trace, logs it, and displays an error page.
func handleInvocationPanic(c *ControllerAgent, err interface{}) {
	log.TRACE.Printf("panic err:%s\n%s", err, string(debug.Stack()))
	c.Result = c.RenderError(NewError(debug.Stack(), err))
}
