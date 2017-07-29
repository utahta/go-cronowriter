package cronowriter

import (
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/lestrrat/go-strftime"
)

type (
	// A CronoWriter writes message to a set of output files.
	CronoWriter struct {
		pattern *strftime.Strftime // given pattern
		path    string             // current file path
		symlink *strftime.Strftime // symbolic link to current file path
		fp      *os.File           // current file pointer
		loc     *time.Location
		mux     sync.Locker
		debug   logger
		init    bool // if true, open the file when New() method is called
	}

	// A Option with CronoWriter.
	Option func(*CronoWriter)
)

var (
	_   io.WriteCloser = (*CronoWriter)(nil) // check if object implements interface
	now                = time.Now            // for test
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
		symlink: nil,
		fp:      nil,
		loc:     time.Local,
		mux:     new(sync.Mutex), // default mutex enable
		debug:   &nopLogger{},
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

// WithLocation set the location to loc.
func WithLocation(loc *time.Location) Option {
	return func(c *CronoWriter) {
		c.loc = loc
	}
}

// WithSymlink enables its creates a symbolic link to the specify pattern.
func WithSymlink(pattern string) Option {
	return func(c *CronoWriter) {
		p, err := strftime.New(pattern)
		if err != nil {
			panic(err)
		}
		c.symlink = p
	}
}

// WithMutex enables its uses sync.Mutex when file writing.
func WithMutex() Option {
	return func(c *CronoWriter) {
		c.mux = new(sync.Mutex)
	}
}

// WithNopMutex disables its uses sync.Mutex when file writing.
func WithNopMutex() Option {
	return func(c *CronoWriter) {
		c.mux = new(nopMutex)
	}
}

// WithDebug enables output stdout and stderr.
func WithDebug() Option {
	return func(c *CronoWriter) {
		c.debug = newDebugLogger()
	}
}

// WithInit enables its creates output file when CronoWriter initialize.
func WithInit() Option {
	return func(c *CronoWriter) {
		c.init = true
	}
}

// Write writes to the file and rotate files automatically based on current date and time.
func (c *CronoWriter) Write(b []byte) (int, error) {
	c.mux.Lock()
	defer c.mux.Unlock()

	t := now().In(c.loc)
	path := c.pattern.FormatString(t)

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
		c.createSymlink(t, path)

		c.path = path
		c.fp = fp
	}

	return c.write(b, nil)
}

func (c *CronoWriter) createSymlink(t time.Time, path string) {
	if c.symlink == nil {
		return
	}

	symlink := c.symlink.FormatString(t)
	if symlink == path {
		c.debug.Error("Can't create symlink. same path is specified.")
		return // ignore error
	}

	if _, err := os.Stat(symlink); err == nil {
		if err := os.Remove(symlink); err != nil {
			c.debug.Error(err)
			return // ignore error
		}
	}

	if err := os.Symlink(path, symlink); err != nil {
		c.debug.Error(err)
		return // ignore error
	}
}

// Close closes file.
func (c *CronoWriter) Close() error {
	c.mux.Lock()
	defer c.mux.Unlock()

	return c.fp.Close()
}

func (c *CronoWriter) write(b []byte, err error) (int, error) {
	if err != nil {
		c.debug.Error(err)
		return 0, err
	}

	c.debug.Write(b)
	return c.fp.Write(b)
}
