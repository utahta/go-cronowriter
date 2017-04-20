package cronowriter

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
	tmpDir, err := ioutil.TempDir("", "cronowriter")
	if err != nil {
		t.Fatal(err)
	}

	c, _ := New("/path/to/file")
	if c.pattern.Pattern() != "/path/to/file" {
		t.Errorf("Expected pattern file, got %s", c.pattern.Pattern())
	}

	c, _ = New("/%Y/%m/%d/%H/%M/%S/file")
	if c.pattern.Pattern() != "/%Y/%m/%d/%H/%M/%S/file" {
		t.Errorf("Expected pattern 2006/01/02/15/04/05/file, got %s", c.pattern.Pattern())
	}

	c, _ = New("/path/to/file", WithLocation(time.UTC))
	if c.loc != time.UTC {
		t.Errorf("Expected location UTC, got %v", c.loc)
	}

	c, _ = New("/path/to/file", WithMutex())
	if c.mux == nil {
		t.Error("Expected mutex object, got nil")
	}

	c, err = New("/path/to/%")
	if err == nil {
		t.Errorf("Expected failed compile error, got %v", err)
	}

	initPath := filepath.Join(tmpDir, "init_test.log")
	_, err = New(initPath, WithInit())
	if err != nil {
		t.Error(err)
	}
	if _, err := os.Stat(initPath); err != nil {
		t.Error(err)
	}
}

func TestMustNew_Panic(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected get panic")
		}
	}()

	MustNew("/path/to/%")
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
		{filepath.Join("2006", "01", "02", "test.log"), filepath.Join("2006", "01", "02", "test.log")},
	}

	jst := time.FixedZone("Asia/Tokyp", 9*60*60)
	for _, test := range tests {
		c := MustNew(filepath.Join(tmpDir, test.pattern), WithLocation(jst))
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

	c := MustNew(filepath.Join(tmpDir, "test.log.%Y%m%d%H%M%S"))
	for _, test := range tests {
		stubNow(test.value)
		if _, err := c.Write([]byte("test")); err != nil {
			t.Fatal(err)
		}
	}
}

func TestCronoWriter_WriteMutex(t *testing.T) {
	tmpDir, err := ioutil.TempDir("", "cronowriter")
	if err != nil {
		t.Fatal(err)
	}
	stubNow("2017-02-04 16:35:05 +0900")

	c := MustNew(filepath.Join(tmpDir, "test.log.%Y%m%d%H%M%S"), WithMutex())
	for i := 0; i < 10; i++ {
		go func() {
			if _, err := c.Write([]byte("test")); err != nil {
				t.Fatal(err)
			}
		}()
	}
}

func TestCronoWriter_Close(t *testing.T) {
	c := MustNew("file")
	if err := c.Close(); err != os.ErrInvalid {
		t.Error(err)
	}
}
