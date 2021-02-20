package bucket

import (
	"io"
	"os"
)

type BucketCloserIface interface {
	BucketIface

	io.Closer
}

type CloserHelper interface {
	// IsClosed returns true is interface is closed.
	// It is helper function (maybe is not need and in future will be removed).
	IsClosed() bool
}

// WithCloser returns input interface with io.Closer and with check of closing.
func WithCloser(in BucketIface, sets ...optCloser) BucketCloserIface {
	opt := closerOpts{}
	for _, set := range sets {
		set(&opt)
	}

	return &wrapCloser{BucketIface: in, closerOpts: opt}
}

type wrapCloser struct {
	BucketIface
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
	return w.BucketIface.ReadAt(p, off)
}

// Reset implements io.Reader interfaces with closing check.
func (w *wrapCloser) Read(p []byte) (n int, err error) {
	if w.closed {
		return 0, os.ErrClosed
	}
	return w.BucketIface.Read(p)
}

// Reset implements io.Writer interfaces with closing check.
func (w *wrapCloser) Write(p []byte) (n int, err error) {
	if w.closed {
		return 0, os.ErrClosed
	}
	w.changed = true
	return w.BucketIface.Write(p)
}

// Reset implements io.WriterAt interfaces with closing check.
func (w *wrapCloser) WriteAt(p []byte, off int64) (n int, err error) {
	if w.closed {
		return 0, os.ErrClosed
	}
	w.changed = true
	return w.BucketIface.WriteAt(p, off)
}

// Reset implements Truncater interfaces with closing check.
func (w *wrapCloser) Truncate(size int64) error {
	if w.closed {
		return os.ErrClosed
	}
	w.changed = true
	return w.BucketIface.Truncate(size)
}

// Reset implements Truncater interfaces with closing check.
func (w *wrapCloser) Reset() {
	if w.closed {
		return
	}
	w.changed = true
	w.BucketIface.Reset()
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

// IsClosed implements CloserHelper interface.
func (w *wrapCloser) IsClosed() bool {
	return w.closed
}
