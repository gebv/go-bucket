package bucket

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func ExampleWithCloser() {
	var seq = 0
	got, want := []byte{}, []byte("foo bar")
	callback := func(changed bool, i BucketIface) (err error) {
		seq++
		// required refresh cursor positio because we written and create buffer without os.O_APPEND
		i.Seek(0, io.SeekStart)
		got, err = ioutil.ReadAll(i)
		return err
	}
	// your buffer
	b := New([]byte{})
	// add for buffer io.Closer with callback
	bc := WithCloser(b, CloseHook(callback))
	// something is written
	fmt.Fprint(bc, "foo")
	fmt.Fprint(bc, " ")
	fmt.Fprint(bc, "bar")
	bc.Close()

	fmt.Println("Seq:", seq)
	fmt.Println("Want:", string(want))
	fmt.Println("Got:", string(got))

	// Output:
	// Seq: 1
	// Want: foo bar
	// Got: foo bar
}

func TestCloser(t *testing.T) {
	t.Run("not closed", func(t *testing.T) {
		c := WithCloser(New([]byte{}, SetAppend(true)))
		assert.False(t, c.(CloserHelper).IsClosed())
		_, err := c.Write([]byte("abc"))
		assert.NoError(t, err)
		_, err = c.WriteAt([]byte("abc"), 3)
		assert.NoError(t, err)
		err = c.Truncate(3)
		assert.NoError(t, err)
		res, err := ioutil.ReadAll(c)
		assert.NoError(t, err)
		assert.EqualValues(t, "abc", string(res))
	})
	t.Run("closed", func(t *testing.T) {
		c := WithCloser(New([]byte{}, SetAppend(true)))
		assert.False(t, c.(CloserHelper).IsClosed())
		err := c.Close()
		assert.NoError(t, err)
		assert.True(t, c.(CloserHelper).IsClosed())
		err = c.Close()
		assert.EqualError(t, os.ErrClosed, err.Error())
		assert.True(t, c.(CloserHelper).IsClosed())

		_, err = c.Write([]byte("abc"))
		assert.EqualError(t, os.ErrClosed, err.Error())
		_, err = c.WriteAt([]byte("abc"), 3)
		assert.EqualError(t, os.ErrClosed, err.Error())
		err = c.Truncate(3)
		assert.EqualError(t, os.ErrClosed, err.Error())
		res, err := ioutil.ReadAll(c)
		assert.EqualError(t, os.ErrClosed, err.Error())
		assert.EqualValues(t, "", string(res))
	})
	t.Run("close hook OK", func(t *testing.T) {
		var seq = 0
		closeFn := func(changed bool, i BucketIface) error {
			seq++
			return nil
		}
		c := WithCloser(New([]byte{}, SetAppend(true)), CloseHook(closeFn))
		assert.False(t, c.(CloserHelper).IsClosed())

		assert.EqualValues(t, 0, seq)
		err := c.Close()
		assert.NoError(t, err)
		assert.True(t, c.(CloserHelper).IsClosed())
		assert.EqualValues(t, 1, seq)
		err = c.Close()
		assert.EqualError(t, os.ErrClosed, err.Error())
		assert.True(t, c.(CloserHelper).IsClosed())
		assert.EqualValues(t, 1, seq)
	})
	t.Run("close hook Err", func(t *testing.T) {
		var seq = 0
		var someErr = fmt.Errorf("some errors")
		closeFn := func(changed bool, i BucketIface) error {
			seq++
			return someErr
		}
		c := WithCloser(New([]byte{}, SetAppend(true)), CloseHook(closeFn))
		assert.False(t, c.(CloserHelper).IsClosed())

		assert.EqualValues(t, 0, seq)
		err := c.Close()
		assert.EqualError(t, someErr, err.Error())
		assert.EqualValues(t, 1, seq)
		assert.True(t, c.(CloserHelper).IsClosed())
		err = c.Close()
		assert.EqualError(t, os.ErrClosed, err.Error())
		assert.True(t, c.(CloserHelper).IsClosed())
		assert.EqualValues(t, 1, seq)
	})
}
