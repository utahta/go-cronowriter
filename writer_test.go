package cronowriter

import (
	"errors"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"
)

var tmpDir string

func TestMain(m *testing.M) {
	var err error
	tmpDir, err = ioutil.TempDir("", "cronowriter")
	if err != nil {
		panic(err)
	}

	os.Exit(m.Run())
}

func stubNow(value string) {
	now = func() time.Time {
		t, _ := time.Parse("2006-01-02 15:04:05 -0700", value)
		return t
	}
}

func StubNow(value string) {
	stubNow(value)
}

func TestNew(t *testing.T) {
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

	c, _ = New("/path/to/file", WithSymlink("/path/to/symlink"))
	if c.symlink.Pattern() != "/path/to/symlink" {
		t.Errorf("Expected symlink pattern /path/to/symlink, got %v", c.loc)
	}

	c, _ = New("/path/to/file", WithMutex())
	if _, ok := c.mux.(*sync.Mutex); !ok {
		t.Errorf("Expected mutex object, got %#v", c.mux)
	}

	c, _ = New("/path/to/file", WithNopMutex())
	if _, ok := c.mux.(*nopMutex); !ok {
		t.Errorf("Expected nop mutex object, got %#v", c.mux)
	}

	c, _ = New("/path/to/file", WithDebug())
	if _, ok := c.log.(*debugLogger); !ok {
		t.Errorf("Expected debugLogger object, got %#v", c.log)
	}

	c, _ = New("/path/to/file", WithStdout())
	if _, ok := c.log.(*stdoutLogger); !ok {
		t.Errorf("Expected stdoutLogger object, got %#v", c.log)
	}

	c, _ = New("/path/to/file", WithStderr())
	if _, ok := c.log.(*stderrLogger); !ok {
		t.Errorf("Expected stderrLogger object, got %#v", c.log)
	}

	c, err := New("/path/to/%")
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
	tests := []struct {
		pattern        string
		expectedSuffix string
	}{
		{"test.log.%Y%m%d%H%M%S", "test.log.20170204163505"},
		{filepath.Join("%Y", "%m", "%d", "test.log"), filepath.Join("2017", "02", "04", "test.log")},
		{filepath.Join("2006", "01", "02", "test.log"), filepath.Join("2006", "01", "02", "test.log")},
		{filepath.Join("2006", "01", "02", "test.log"), filepath.Join("2006", "01", "02", "test.log")}, // repeat
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

	expectText := "hello symlink"
	c := MustNew(filepath.Join(tmpDir, "test.log"), WithSymlink(filepath.Join(tmpDir, "test-symlink.log")))
	if _, err := c.Write([]byte(expectText)); err != nil {
		t.Fatal(err)
	}

	b, err := ioutil.ReadFile(filepath.Join(tmpDir, "test-symlink.log"))
	if err != nil {
		t.Fatal(err)
	}

	if string(b) != expectText {
		t.Errorf("Expected %s, got %s", expectText, string(b))
	}
}

func TestCronoWriter_Path(t *testing.T) {
	stubNow("2017-02-04 16:35:05 +0900")

	jst := time.FixedZone("Asia/Tokyp", 9*60*60)
	c := MustNew(filepath.Join(tmpDir, "test.log.%Y%m%d%H%M%S"), WithInit(), WithLocation(jst))
	expected := filepath.Join(tmpDir, "test.log.20170204163505")
	if expected != c.Path() {
		t.Errorf("Expected file path %s, got %s", expected, c.Path())
	}
}

func TestCronoWriter_WriteSymlink(t *testing.T) {
	stubNow("2017-02-04 16:35:05 +0900")
	tests := []struct {
		pattern      string
		expectedText string
	}{
		{"test.log.1", "hello symlink"},
		{"test.log.1", "hello symlinkhello symlink"},
		{"test.log.2", "hello symlink"},
	}

	for _, test := range tests {
		sympath := filepath.Join(tmpDir, "test-symlink.log")
		c := MustNew(filepath.Join(tmpDir, test.pattern), WithSymlink(sympath))
		if _, err := c.Write([]byte("hello symlink")); err != nil {
			t.Fatal(err)
		}

		b, err := ioutil.ReadFile(sympath)
		if err != nil {
			t.Fatal(err)
		}

		if string(b) != test.expectedText {
			t.Errorf("Expected %s, got %s", test.expectedText, string(b))
		}
	}

	samepath := filepath.Join(tmpDir, "test.log")
	c := MustNew(samepath, WithSymlink(samepath))
	if _, err := c.Write([]byte("hello symlink")); err != nil {
		t.Fatal(err)
	}
}

func TestCronoWriter_WriteRepeat(t *testing.T) {
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
	stubNow("2017-02-04 16:35:05 +0900")

	c := MustNew(filepath.Join(tmpDir, "test.log.%Y%m%d%H%M%S"), WithMutex())
	wg := &sync.WaitGroup{}
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if _, err := c.Write([]byte("test")); err != nil {
				t.Fatal(err)
			}
		}()
	}
	wg.Wait()
}

func TestCronoWriter_Write_MkdirFail(t *testing.T) {
	c := MustNew(filepath.Join("/cronowriter", "test.log"))
	if _, err := c.Write([]byte("test")); err == nil {
		t.Error("Expected permission denied error, got nil")
	}
}

func TestCronoWriter_Write_OpenfileFail(t *testing.T) {
	c := MustNew(filepath.Join("/", "test.log"))
	if _, err := c.Write([]byte("test")); err == nil {
		t.Error("Expected permission denied error, got nil")
	}
}

func TestCronoWriter_Write_SymlinkFail(t *testing.T) {
	path := filepath.Join("/cronowriter", "test.log")
	c := MustNew(filepath.Join(tmpDir, "test.log"), WithSymlink(path))
	c.Write([]byte("test"))

	if _, err := os.Stat(path); err == nil {
		t.Errorf("Expected no such file or directory error, got %v", err)
	}
}

func TestCronoWriter_write(t *testing.T) {
	c := MustNew("/path/to/test.log")
	expectedErr := errors.New("fail write")

	n, err := c.write(nil, expectedErr)
	if n != 0 {
		t.Errorf("Expected write len 0, got %v", n)
	}
	if err != expectedErr {
		t.Errorf("Expected error %v, got %v", expectedErr, err)
	}
}

func TestCronoWriter_Close(t *testing.T) {
	c := MustNew("file")
	if err := c.Close(); err != os.ErrInvalid {
		t.Error(err)
	}
}
