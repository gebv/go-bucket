package bucket

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"testing"
)

func TestBucketHappyPaths(t *testing.T) {
	regularChecks := func(t *testing.T, b *Bucket, wantData []byte, wantOff int) {
		t.Helper()
		equalBytes(t, wantData, b.data)
		equalInts(t, int(wantOff), int(b.off))
	}
	t.Run("writeTo_presetWRAppend", func(t *testing.T) {
		b := NewBucket([]byte("abc"), os.O_RDWR|os.O_APPEND)
		fmt.Fprint(b, "def")
		_, err := b.Write([]byte("ghi"))
		noError(t, err)
		regularChecks(t, b, []byte("abcdefghi"), 0)
	})
	t.Run("writeTo_presetWR", func(t *testing.T) {
		b := NewBucket([]byte("abc"), os.O_RDWR)
		fmt.Fprint(b, "def")
		_, err := b.Write([]byte("ghi"))
		noError(t, err)
		regularChecks(t, b, []byte("defghi"), 6)
	})
	t.Run("writeTo_presetWriteOnly", func(t *testing.T) {
		b := NewBucket([]byte("abc"), os.O_WRONLY)
		fmt.Fprint(b, "def")
		_, err := b.Write([]byte("ghi"))
		noError(t, err)
		regularChecks(t, b, []byte("defghi"), 6)
	})
	t.Run("writeTo_presetWriteOnlyAppend", func(t *testing.T) {
		b := NewBucket([]byte("abc"), os.O_WRONLY|os.O_APPEND)
		fmt.Fprint(b, "def")
		_, err := b.Write([]byte("ghi"))
		noError(t, err)
		regularChecks(t, b, []byte("abcdefghi"), 0)
	})
	t.Run("writeTo_presetReadOnly", func(t *testing.T) {
		b := NewBucket([]byte("abc"), os.O_RDONLY)
		fmt.Fprint(b, "def")
		_, err := b.Write([]byte("def"))
		equalError(t, ErrNoAccess, err)
		regularChecks(t, b, []byte("abc"), 0)
	})

	t.Run("writeAndRead_presetRWAppend", func(t *testing.T) {
		b := NewBucket([]byte("abc"), os.O_RDWR|os.O_APPEND)
		fmt.Fprint(b, "def")
		regularChecks(t, b, []byte("abcdef"), 0)

		got, err := ioutil.ReadAll(b)
		noError(t, err)
		equalBytes(t, []byte("abcdef"), got)
		regularChecks(t, b, []byte("abcdef"), 6)

		got, err = ioutil.ReadAll(b)
		noError(t, err)
		equalBytes(t, []byte(""), got)
		regularChecks(t, b, []byte("abcdef"), 6)
	})

	t.Run("seeker_presetRWAppend", func(t *testing.T) {
		b := NewBucket([]byte("abc"), os.O_RDWR|os.O_APPEND)
		fmt.Fprint(b, "def")
		regularChecks(t, b, []byte("abcdef"), 0)

		got, err := ioutil.ReadAll(b)
		noError(t, err)
		equalBytes(t, []byte("abcdef"), got)
		regularChecks(t, b, []byte("abcdef"), 6)

		_, err = b.Seek(0, io.SeekStart)
		noError(t, err)

		got, err = ioutil.ReadAll(b)
		noError(t, err)
		equalBytes(t, []byte("abcdef"), got)
		regularChecks(t, b, []byte("abcdef"), 6)

		_, err = b.Seek(3, io.SeekStart)
		noError(t, err)

		got, err = ioutil.ReadAll(b)
		noError(t, err)
		equalBytes(t, []byte("def"), got)
		regularChecks(t, b, []byte("abcdef"), 6)

		_, err = b.Seek(-3, io.SeekCurrent)
		noError(t, err)

		got, err = ioutil.ReadAll(b)
		noError(t, err)
		equalBytes(t, []byte("def"), got)
		regularChecks(t, b, []byte("abcdef"), 6)

		_, err = b.Seek(0, io.SeekEnd)
		noError(t, err)

		got, err = ioutil.ReadAll(b)
		noError(t, err)
		equalBytes(t, []byte(""), got)
		regularChecks(t, b, []byte("abcdef"), 6)
	})

	t.Run("trunReset_presetRWAppend", func(t *testing.T) {
		b := NewBucket([]byte("abc"), os.O_RDWR|os.O_APPEND)
		fmt.Fprint(b, "def")
		regularChecks(t, b, []byte("abcdef"), 0)

		b.Truncate(3)
		regularChecks(t, b, []byte("abc"), 0)

		b.Reset()

		regularChecks(t, b, []byte(""), 0)
	})

	t.Run("trunReset_presetRW", func(t *testing.T) {
		b := NewBucket([]byte("abc"), os.O_RDWR)
		fmt.Fprint(b, "def")
		regularChecks(t, b, []byte("def"), 3)

		b.Truncate(3)
		regularChecks(t, b, []byte("def"), 3)

		b.Reset()

		regularChecks(t, b, []byte(""), 0)
	})

	t.Run("closed", func(t *testing.T) {
		b := NewBucket([]byte("abc"), os.O_RDWR)
		fmt.Fprint(b, "def")
		regularChecks(t, b, []byte("def"), 3)

		err := b.Close()
		noError(t, err)

		err = b.Close()
		equalError(t, os.ErrClosed, err)

		_, err = b.Read(make([]byte, 3))
		equalError(t, os.ErrClosed, err)

		_, err = b.Write(make([]byte, 3))
		equalError(t, os.ErrClosed, err)

		err = b.Truncate(3)
		equalError(t, os.ErrClosed, err)
	})

	t.Run("read_EOF", func(t *testing.T) {
		b := NewBucket([]byte("abc"), os.O_RDWR|os.O_APPEND)
		fmt.Fprint(b, "def")
		regularChecks(t, b, []byte("abcdef"), 0)

		_, err := b.ReadAt(make([]byte, 3), b.Size())
		equalError(t, io.EOF, err)
		_, err = b.ReadAt(make([]byte, 3), b.Size()+1)
		equalError(t, io.ErrUnexpectedEOF, err)

		_, err = b.Read(make([]byte, 3))
		noError(t, err)

		_, err = b.ReadAt(make([]byte, 3), b.Size())
		equalError(t, io.EOF, err)
		_, err = b.ReadAt(make([]byte, 3), b.Size()+1)
		equalError(t, io.ErrUnexpectedEOF, err)

		_, err = b.Read(make([]byte, 3))
		noError(t, err)

		_, err = b.Read(make([]byte, 3))
		equalError(t, io.EOF, err)
		_, err = b.ReadAt(make([]byte, 3), b.Size())
		equalError(t, io.EOF, err)

		_, err = b.ReadAt(make([]byte, 3), b.Size()+1)
		equalError(t, io.ErrUnexpectedEOF, err)
	})

	t.Run("readOnly", func(t *testing.T) {
		b := NewBucket([]byte("abc"), os.O_RDONLY)
		fmt.Fprint(b, "def")
		regularChecks(t, b, []byte("abc"), 0)

		err := b.Truncate(3)
		equalError(t, ErrNoAccess, err)
		err = b.Truncate(12)
		equalError(t, ErrNoAccess, err)
		err = b.Truncate(0)
		equalError(t, ErrNoAccess, err)
	})

	t.Run("writeOnlyAndTruncate", func(t *testing.T) {
		b := NewBucket([]byte("abc"), os.O_WRONLY)
		fmt.Fprint(b, "def")
		regularChecks(t, b, []byte("def"), 3)

		err := b.Truncate(3)
		noError(t, err)
		regularChecks(t, b, []byte("def"), 3)
		equalInts(t, int(b.Cap()), 3)

		err = b.Truncate(6)
		noError(t, err)
		regularChecks(t, b, []byte("def\x00\x00\x00"), 3)
		equalInts(t, int(b.Cap()), 6)

		err = b.Truncate(0)
		noError(t, err)
		regularChecks(t, b, []byte(""), 0)

		_, err = b.Read(make([]byte, 3))
		equalError(t, ErrNoAccess, err)

		_, err = b.ReadAt(make([]byte, 3), 3)
		equalError(t, ErrNoAccess, err)
	})

	t.Run("ErrOutOfRange", func(t *testing.T) {
		b := NewBucket([]byte("abc"), os.O_RDWR|os.O_APPEND)
		fmt.Fprint(b, "def")
		regularChecks(t, b, []byte("abcdef"), 0)

		err := b.Truncate(-1)
		equalError(t, ErrNegativeSize, err)
	})

	// TODO: more tests for seeker
}

func equalBytes(t *testing.T, want, got []byte) {
	t.Helper()
	if !bytes.Equal(want, got) {
		t.Errorf("Should be equal data, but: want=%s != got=%s", want, got)
	}
}

func equalInts(t *testing.T, want, got int) {
	t.Helper()
	if want != got {
		t.Errorf("Should be equal ints, but: want=%v != got=%v", want, got)
	}
}

func noError(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatal("Should be no error, but got:", err)
	}
}

func equalError(t *testing.T, want, got error) {
	t.Helper()
	if want == nil {
		t.Fatal("Should be equal errors and does is not nil, want is nil (incorrect use of the checker, please choose another)")
	}
	if got == nil {
		t.Error("Should be equal errors and does is not nil, got is nil")
		return
	}
	if want != got {
		t.Errorf("Should be equal errors, but: want=%v != got=%v", want, got)
	}
}
