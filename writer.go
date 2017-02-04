package writer

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type cronoWriter struct {
	baseDir string
	pattern string
	path    string
	fp      *os.File
	loc     *time.Location
}

type option func(*cronoWriter)

var (
	_   io.WriteCloser   = &cronoWriter{} // check if object implements interface
	now func() time.Time = time.Now       // for test
)

// New returns the cronoWriter
func New(baseDir, pattern string, options ...option) *cronoWriter {
	pattern = replacePattern(pattern)

	c := &cronoWriter{
		baseDir: baseDir,
		pattern: pattern,
		path:    "",
		fp:      nil,
		loc:     time.Local,
	}

	for _, option := range options {
		option(c)
	}
	return c
}

func replacePattern(p string) string {
	p = strings.Replace(p, "%Y", "2006", -1)
	p = strings.Replace(p, "%m", "01", -1)
	p = strings.Replace(p, "%d", "02", -1)
	p = strings.Replace(p, "%H", "15", -1)
	p = strings.Replace(p, "%M", "04", -1)
	p = strings.Replace(p, "%S", "05", -1)
	return p
}

func WithLocation(loc *time.Location) option {
	return func(c *cronoWriter) {
		c.loc = loc
	}
}

func (c *cronoWriter) Write(b []byte) (int, error) {
	path := filepath.Join(c.baseDir, now().In(c.loc).Format(c.pattern))

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
	return c.fp.Close()
}
