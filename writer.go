package writer

import (
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/lestrrat/go-strftime"
)

type CronoWriter struct {
	pattern *strftime.Strftime // given pattern
	path    string             // current file path
	fp      *os.File           // current file pointer
	loc     *time.Location
	mux     sync.Locker
	stdout  io.Writer
	stderr  io.Writer
	init    bool // if true, open the file when New() method is called
}

type Option func(*CronoWriter)

var (
	_   io.WriteCloser   = &CronoWriter{} // check if object implements interface
	now func() time.Time = time.Now       // for test
)

// New returns a CronoWriter with the given pattern and options.
func New(pattern string, options ...Option) (*CronoWriter, error) {
	p, err := strftime.New(pattern)
	if err != nil {
		return nil, err
	}

	c := &CronoWriter{
		pattern: p,
		path:    "",
		fp:      nil,
		loc:     time.Local,
		mux:     new(noopMutex), // default mutex off
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
func MustNew(pattern string, options ...Option) *CronoWriter {
	c, err := New(pattern, options...)
	if err != nil {
		panic(err)
	}
	return c
}

func WithLocation(loc *time.Location) Option {
	return func(c *CronoWriter) {
		c.loc = loc
	}
}

func WithMutex() Option {
	return func(c *CronoWriter) {
		c.mux = new(sync.Mutex)
	}
}

func WithDebug() Option {
	return func(c *CronoWriter) {
		c.stdout = os.Stdout
		c.stderr = os.Stderr
	}
}

func WithInit() Option {
	return func(c *CronoWriter) {
		c.init = true
	}
}

func (c *CronoWriter) Write(b []byte) (int, error) {
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

func (c *CronoWriter) Close() error {
	c.mux.Lock()
	defer c.mux.Unlock()

	return c.fp.Close()
}

func (c *CronoWriter) write(b []byte, err error) (int, error) {
	if err != nil {
		c.stderr.Write([]byte(err.Error()))
		return 0, err
	}

	c.stdout.Write(b)
	return c.fp.Write(b)
}
