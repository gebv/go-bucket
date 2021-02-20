package bucket

import (
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCloser(t *testing.T) {
	t.Run("not closed", func(t *testing.T) {
		c := WithCloser(New([]byte{}, SetAppend(true)))
		assert.False(t, c.IsClosed())
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
		assert.False(t, c.IsClosed())
		err := c.Close()
		assert.NoError(t, err)
		assert.True(t, c.IsClosed())
		err = c.Close()
		assert.EqualError(t, os.ErrClosed, err.Error())
		assert.True(t, c.IsClosed())

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
		closeFn := func(changed bool, i withCloserInput) error {
			seq++
			return nil
		}
		c := WithCloser(New([]byte{}, SetAppend(true)), CloseHook(closeFn))
		assert.False(t, c.IsClosed())

		err := c.Close()
		assert.NoError(t, err)
		assert.True(t, c.IsClosed())
		err = c.Close()
		assert.EqualError(t, os.ErrClosed, err.Error())
		assert.True(t, c.IsClosed())

		assert.EqualValues(t, 1, seq)
	})
	t.Run("close hook Err", func(t *testing.T) {
		var seq = 0
		var someErr = fmt.Errorf("some errors")
		closeFn := func(changed bool, i withCloserInput) error {
			seq++
			return someErr
		}
		c := WithCloser(New([]byte{}, SetAppend(true)), CloseHook(closeFn))
		assert.False(t, c.IsClosed())

		err := c.Close()
		assert.EqualError(t, someErr, err.Error())
		assert.True(t, c.IsClosed())
		err = c.Close()
		assert.EqualError(t, os.ErrClosed, err.Error())
		assert.True(t, c.IsClosed())

		assert.EqualValues(t, 1, seq)
	})
}