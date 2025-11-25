package test

import (
	"fmt"
	"testing"

	"github.com/byte-wright/expect"
)

// Logger implements Test and allows for inspection of the
// calls.
type Logger struct {
	Fatals   []string
	Errors   []string
	Messages []string
	t        *testing.T
}

func New(t *testing.T, f func(t expect.Test)) (logger *Logger) {
	t.Helper()

	logger = &Logger{t: t}

	defer func() {
		if r := recover(); r != nil {
			fmt.Println("recovered panic", r)
		}
	}()

	f(logger)

	return logger
}

// Fatalf records call.
func (l *Logger) Fatalf(f string, i ...interface{}) {
	line := fmt.Sprintf(f, i...)
	l.Fatals = append(l.Fatals, line)
	l.Messages = append(l.Messages, line)

	panic("fatal")
}

// Errorf records call.
func (l *Logger) Errorf(f string, i ...interface{}) {
	line := fmt.Sprintf(f, i...)
	l.Errors = append(l.Errors, line)
	l.Messages = append(l.Messages, line)
}

// Error records call.
func (l *Logger) Error(p ...interface{}) {
	line := fmt.Sprint(p...)
	l.Errors = append(l.Errors, line)
	l.Messages = append(l.Messages, line)
}

func (l *Logger) Helper() {
	// can be ignored, error logger does not care about location information
}

// ExpectMessages returns the messages as a expect-value.
func (l *Logger) ExpectMessages() expect.Val {
	return expect.Value(l.t, "messages", l.Messages)
}

// ExpectMessage returns the message at given index.
func (l *Logger) ExpectMessage(i int) expect.Val {
	l.t.Helper()
	if len(l.Messages) <= i {
		l.t.Errorf("there is no message at index %v", i)
		return expect.Value(l.t, "message", "")
	}

	return expect.Value(l.t, "message", l.Messages[i])
}
