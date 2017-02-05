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
	pattern *strftime.Strftime
	path    string
	fp      *os.File
	loc     *time.Location
	mux     sync.Locker
}

type option func(*cronoWriter)

var (
	_                 io.WriteCloser   = &cronoWriter{} // check if object implements interface
	now               func() time.Time = time.Now       // for test
	waitCloseDuration                  = 5 * time.Second
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
	}

	for _, option := range options {
		option(c)
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
			time.Sleep(waitCloseDuration) // any ideas?
			fp.Close()
		}(c.fp)

		dir := filepath.Dir(path)
		if err := os.MkdirAll(dir, os.ModePerm); err != nil {
			return 0, err
		}

		fp, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
		if err != nil {
			return 0, err
		}
		c.path = path
		c.fp = fp
	}
	return c.fp.Write(b)
}

func (c *cronoWriter) Close() error {
	c.mux.Lock()
	defer c.mux.Unlock()

	return c.fp.Close()
}
