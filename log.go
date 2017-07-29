package cronowriter

import (
	"fmt"
	"io"
	"os"
)

type (
	logger interface {
		Write(b []byte)
		Error(args ...interface{})
		Errorf(format string, args ...interface{})
	}

	nopLogger struct{}

	debugLogger struct {
		stdout io.Writer
		stderr io.Writer
	}
)

func newDebugLogger() *debugLogger {
	return &debugLogger{
		stdout: os.Stdout,
		stderr: os.Stderr,
	}
}

func (l *nopLogger) Write(b []byte)                            {}
func (l *nopLogger) Error(args ...interface{})                 {}
func (l *nopLogger) Errorf(format string, args ...interface{}) {}

func (l *debugLogger) Write(b []byte) {
	fmt.Fprintf(l.stdout, "%s", b)
}

func (l *debugLogger) Error(args ...interface{}) {
	fmt.Fprintln(l.stderr, args...)
}

func (l *debugLogger) Errorf(format string, args ...interface{}) {
	fmt.Fprintf(l.stderr, format, args...)
}
