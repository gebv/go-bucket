package bucket

import (
	"io"
	"os"
)

type withCloserInput interface {
	RatWatSeeker
	Truncater
}

type withCloserOutput interface {
	RatWatSeeker
	Truncater
	io.Closer

	// IsClosed returns true is interface is closed.
	// It is helper function (maybe is not need and in future will be removed).
	IsClosed() bool
}

// WithCloser returns input interface with io.Closer and with check of closing.
func WithCloser(in withCloserInput, sets ...optCloser) withCloserOutput {
	opt := closerOpts{}
	for _, set := range sets {
		set(&opt)
	}

	return &wrapCloser{withCloserInput: in, closerOpts: opt}
}

type wrapCloser struct {
	withCloserInput
	// indicates if the interface has been closed
	closed bool
	// indicates if the data has changed
	changed bool

	closerOpts
}

// Reset implements io.ReaderAt interfaces with closing check.
func (w *wrapCloser) ReadAt(p []byte, off int64) (n int, err error) {
	if w.closed {
		return 0, os.ErrClosed
	}
	return w.withCloserInput.ReadAt(p, off)
}

// Reset implements io.Reader interfaces with closing check.
func (w *wrapCloser) Read(p []byte) (n int, err error) {
	if w.closed {
		return 0, os.ErrClosed
	}
	return w.withCloserInput.Read(p)
}

// Reset implements io.Writer interfaces with closing check.
func (w *wrapCloser) Write(p []byte) (n int, err error) {
	if w.closed {
		return 0, os.ErrClosed
	}
	w.changed = true
	return w.withCloserInput.Write(p)
}

// Reset implements io.WriterAt interfaces with closing check.
func (w *wrapCloser) WriteAt(p []byte, off int64) (n int, err error) {
	if w.closed {
		return 0, os.ErrClosed
	}
	w.changed = true
	return w.withCloserInput.WriteAt(p, off)
}

// Reset implements Truncater interfaces with closing check.
func (w *wrapCloser) Truncate(size int64) error {
	if w.closed {
		return os.ErrClosed
	}
	w.changed = true
	return w.withCloserInput.Truncate(size)
}

// Reset implements Truncater interfaces with closing check.
func (w *wrapCloser) Reset() {
	if w.closed {
		return
	}
	w.changed = true
	w.withCloserInput.Reset()
}

// Reset implements io.Closer interfaces with closing check.
func (w *wrapCloser) Close() error {
	if w.closed {
		return os.ErrClosed
	}
	defer func() { w.closed = true }()
	if w.closeFn != nil {
		return w.closeFn(w.changed, w)
	}
	return nil
}

func (w *wrapCloser) IsClosed() bool {
	return w.closed
}
