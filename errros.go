package main

import (
	"errors"
	"fmt"
)

var (
	ErrSignal        = errors.New("received a signal")
	ErrUnsupportedOs = errors.New("this OS is not supported")
)

type execErr struct {
	cmd   string
	msg   string
	cause error
}

func (e *execErr) Error() string {
	return fmt.Sprintf("cmd: %q: %s: Cause: %v", e.cmd, e.msg, e.cause)
}

func (e *execErr) Is(target error) bool {
	t, ok := target.(*execErr)
	if !ok {
		return false
	}

	return t.cmd == e.cmd
}

func (e *execErr) Unwrap() error {
	return e.cause
}
