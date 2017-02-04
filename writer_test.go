package writer

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func stubNow(value string) {
	now = func() time.Time {
		t, _ := time.Parse("2006-01-02 15:04:05 -0700", value)
		return t
	}
}

func TestNew(t *testing.T) {
	c := New("/path/to", "file")
	if c.pattern != "file" {
		t.Errorf("Expected pattern file, got %s", c.pattern)
	}

	c = New("/", "%Y/%m/%d/%H/%M/%S/file")
	if c.pattern != "2006/01/02/15/04/05/file" {
		t.Errorf("Expected pattern 2006/01/02/15/04/05/file, got %s", c.pattern)
	}

	c = New("/path/to", "file", WithLocation(time.UTC))
	if c.loc != time.UTC {
		t.Errorf("Expected location UTC, got %v", c.loc)
	}
}

func TestCronoWriter_Write(t *testing.T) {
	stubNow("2017-02-04 16:35:05 +0900")
	tmpDir, err := ioutil.TempDir("", "cronowriter")
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		pattern        string
		expectedSuffix string
	}{
		{"test.log.%Y%m%d%H%M%S", "test.log.20170204163505"},
		{filepath.Join("%Y", "%m", "%d", "test.log"), filepath.Join("2017", "02", "04", "test.log")},
		{filepath.Join("2006", "01", "02", "test.log"), filepath.Join("2017", "02", "04", "test.log")},
	}

	jst := time.FixedZone("Asia/Tokyp", 9*60*60)
	for _, test := range tests {
		c := New(tmpDir, test.pattern, WithLocation(jst))
		for i := 0; i < 2; i++ {
			if _, err := c.Write([]byte("test")); err != nil {
				t.Fatal(err)
			}
		}

		if _, err := os.Stat(c.path); err != nil {
			t.Fatal(err)
		}

		if !strings.HasSuffix(c.path, test.expectedSuffix) {
			t.Fatalf("Expected suffix %s, got %s", test.expectedSuffix, c.path)
		}
	}
}

func TestCronoWriter_WriteRepeat(t *testing.T) {
	tmpDir, err := ioutil.TempDir("", "cronowriter")
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		value string
	}{
		{"2017-02-04 16:35:05 +0900"},
		{"2017-02-04 16:35:05 +0900"},
		{"2017-02-04 16:35:07 +0900"},
		{"2017-02-04 16:35:08 +0900"},
		{"2017-02-04 16:35:09 +0900"},
	}

	c := New(tmpDir, "test.log.%Y%m%d%H%M%S")
	for _, test := range tests {
		stubNow(test.value)
		if _, err := c.Write([]byte("test")); err != nil {
			t.Fatal(err)
		}
	}
}

func TestCronoWriter_Close(t *testing.T) {
	c := New("", "file")
	if err := c.Close(); err != os.ErrInvalid {
		t.Error(err)
	}
}
