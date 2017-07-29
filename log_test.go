package cronowriter

import (
	"bytes"
	"strings"
	"testing"
)

func TestNopLogger(t *testing.T) {
	l := &nopLogger{}
	l.Write([]byte("test"))
	l.Error("error")
	l.Errorf("error %s", "error")
}

func TestDebugLogger_Write(t *testing.T) {
	obuf := &bytes.Buffer{}
	l := &debugLogger{stdout: obuf}
	outStr := "out text"

	l.Write([]byte(outStr))
	if !strings.Contains(obuf.String(), outStr) {
		t.Errorf("Expected stdout %s, got %s", outStr, obuf.String())
	}
}

func TestDebugLogger_Error(t *testing.T) {
	ebuf := &bytes.Buffer{}
	l := &debugLogger{stderr: ebuf}
	errStr := "err text"

	l.Error(errStr)
	if !strings.Contains(ebuf.String(), errStr) {
		t.Errorf("Expected stderr %s, got %s", errStr, ebuf.String())
	}
}

func TestDebugLogger_Errorf(t *testing.T) {
	ebuf := &bytes.Buffer{}
	l := &debugLogger{stderr: ebuf}
	errStr := "err text"

	l.Errorf("%s", errStr)
	if !strings.Contains(ebuf.String(), errStr) {
		t.Errorf("Expected stderr %s, got %s", errStr, ebuf.String())
	}
}
