package writer

import (
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/lestrrat/go-strftime"
)

type cronoWriter struct {
	pattern *strftime.Strftime // given pattern
	path    string             // current file path
	fp      *os.File           // current file pointer
	loc     *time.Location
	mux     sync.Locker
	stdout  io.Writer
	stderr  io.Writer
	init    bool // if true, open the file when New() method is called
}

type option func(*cronoWriter)

type noopWriter struct{}

func (*noopWriter) Write([]byte) (int, error) {
	return 0, nil // no-op
}

var (
	_   io.WriteCloser   = &cronoWriter{} // check if object implements interface
	now func() time.Time = time.Now       // for test
)

// New returns a cronoWriter with the given pattern and options.
func New(pattern string, options ...option) (*cronoWriter, error) {
	p, err := strftime.New(pattern)
	if err != nil {
		return nil, err
	}

	c := &cronoWriter{
		pattern: p,
		path:    "",
		fp:      nil,
		loc:     time.Local,
		mux:     new(NoMutex), // default mutex off
		stdout:  &noopWriter{},
		stderr:  &noopWriter{},
		init:    false,
	}

	for _, option := range options {
		option(c)
	}

	if c.init {
		if _, err := c.Write([]byte("")); err != nil {
			return nil, err
		}
	}
	return c, nil
}

// MustNew is a convenience function equivalent to New that panics on failure
// instead of returning an error.
func MustNew(pattern string, options ...option) *cronoWriter {
	c, err := New(pattern, options...)
	if err != nil {
		panic(err)
	}
	return c
}

func WithLocation(loc *time.Location) option {
	return func(c *cronoWriter) {
		c.loc = loc
	}
}

func WithMutex() option {
	return func(c *cronoWriter) {
		c.mux = new(sync.Mutex)
	}
}

func WithDebug() option {
	return func(c *cronoWriter) {
		c.stdout = os.Stdout
		c.stderr = os.Stderr
	}
}

func WithInit() option {
	return func(c *cronoWriter) {
		c.init = true
	}
}

func (c *cronoWriter) Write(b []byte) (int, error) {
	c.mux.Lock()
	defer c.mux.Unlock()

	path := c.pattern.FormatString(now().In(c.loc))

	if c.path != path {
		// close file
		go func(fp *os.File) {
			if fp == nil {
				return
			}
			fp.Close()
		}(c.fp)

		dir := filepath.Dir(path)
		if err := os.MkdirAll(dir, os.ModePerm); err != nil {
			return c.write(nil, err)
		}

		fp, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
		if err != nil {
			return c.write(nil, err)
		}
		c.path = path
		c.fp = fp
	}

	return c.write(b, nil)
}

func (c *cronoWriter) Close() error {
	c.mux.Lock()
	defer c.mux.Unlock()

	return c.fp.Close()
}

func (c *cronoWriter) write(b []byte, err error) (int, error) {
	if err != nil {
		c.stderr.Write([]byte(err.Error()))
		return 0, err
	}

	c.stdout.Write(b)
	return c.fp.Write(b)
}
