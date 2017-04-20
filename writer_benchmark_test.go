package cronowriter

import (
	"io/ioutil"
	"path/filepath"
	"testing"
)

func BenchmarkCronoWriter_Write(b *testing.B) {
	stubNow("2017-02-06 20:30:00 +0900")
	tmpDir, err := ioutil.TempDir("", "cronowriter")
	if err != nil {
		b.Fatal(err)
	}

	c := MustNew(filepath.Join(tmpDir, "benchmark.log.%Y%m%d"))
	for i := 0; i < b.N; i++ {
		c.Write([]byte("abcdefg"))
	}
}

func BenchmarkCronoWriter_WriteWithMutex(b *testing.B) {
	stubNow("2017-02-06 20:30:00 +0900")
	tmpDir, err := ioutil.TempDir("", "cronowriter")
	if err != nil {
		b.Fatal(err)
	}

	c := MustNew(filepath.Join(tmpDir, "benchmark.log.%Y%m%d"), WithMutex())
	for i := 0; i < b.N; i++ {
		c.Write([]byte("abcdefg"))
	}
}

func BenchmarkCronoWriter_WriteWithDebug(b *testing.B) {
	stubNow("2017-02-06 20:30:00 +0900")
	tmpDir, err := ioutil.TempDir("", "cronowriter")
	if err != nil {
		b.Fatal(err)
	}

	c := MustNew(filepath.Join(tmpDir, "benchmark.log.%Y%m%d"), WithDebug())
	c.debug.(*debugLogger).stdout = ioutil.Discard
	c.debug.(*debugLogger).stderr = ioutil.Discard
	for i := 0; i < b.N; i++ {
		c.Write([]byte("abcdefg"))
	}
}
