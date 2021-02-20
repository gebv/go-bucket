package bucket

import (
	"errors"
	"io"
)

// RatWatSeeker the name is formed from interfaces io.Reader, io.ReaderAt, io.Writer, io.WriterAt, io.Seeker.
type RatWatSeeker interface {
	io.Reader
	io.ReaderAt
	io.Writer
	io.WriterAt
	io.Seeker
}

type Sizer interface {
	Size() int64
	Cap() int64
}

type Truncater interface {
	Truncate(size int64) error
	Reset()
}

var _ RatWatSeeker = (*Bucket)(nil)
var _ Sizer = (*Bucket)(nil)
var _ Truncater = (*Bucket)(nil)

// New returns bucket with initial data and is append = false.
func New(dat []byte, sets ...optBucket) *Bucket {
	opt := bucketOpts{}
	for _, set := range sets {
		set(&opt)
	}

	return &Bucket{
		data:       dat,
		bucketOpts: opt,
	}
}

// Bucket this is special container store data and implements io.Reader, io.ReaderAt, io.Writer, io.WriterAt, io.Seeker interfaces.
// Shared cursor for writer and reader.
// Implements io.Reader, io.ReaderAt, io.Writer, io.WriterAt, io.Seeker interfaces.
type Bucket struct {
	bucketOpts
	// bucket contents
	data []byte
	// position of cursor for reader and writer
	off int64
}

// Reset resets the bucket to be empty, but capacity is not changes.
func (b *Bucket) Reset() {
	b.data = b.data[:0]
	b.off = 0
}

// Truncate truncates container size data to specified size. Growing if the size exceeds the current data size.
// Continues to use the same allocated storage.
// Returns errors if is closed or if there is no access for the operation.
func (b *Bucket) Truncate(size int64) error {
	if size == 0 {
		b.Reset()
		return nil
	}
	if size < 0 {
		return ErrNegativeSize
	}
	if int(size) > len(b.data) {
		_, err := b.writeAt(make([]byte, int(size)-len(b.data)), size-int64(len(b.data)))
		if err != nil {
			return nil
		}
	}
	b.data = b.data[:size]
	return nil
}

var _ io.Reader = (*Bucket)(nil)

// Read implements the io.Reader interface.
// Cursor is shiftes. Otherwise, the behavior is the same as ReadAt.
func (b *Bucket) Read(buf []byte) (n int, err error) {
	n, err = b.ReadAt(buf, b.off)
	if err != nil {
		return 0, err
	}
	b.off += int64(n)
	return n, nil
}

var _ io.ReaderAt = (*Bucket)(nil)

// ReadAt implements the io.ReaderAt interface.
// Returns error for standard error for the reader.
// Does not change the internal state.
func (b *Bucket) ReadAt(buf []byte, off int64) (n int, err error) {
	if len(buf) > 0 && off == int64(len(b.data)) {
		return 0, io.EOF
	}
	if off > int64(len(b.data)) {
		return 0, io.ErrUnexpectedEOF
	}
	n = copy(buf, b.data[off:])
	return
}

var _ io.Seeker = (*Bucket)(nil)

// Seek implements the io.Seeker interface.
func (b *Bucket) Seek(offset int64, whence int) (int64, error) {
	var abs int64
	switch whence {
	case io.SeekStart:
		abs = offset
	case io.SeekCurrent:
		abs = b.off + offset
	case io.SeekEnd:
		abs = int64(len(b.data)) + offset
	default:
		return 0, ErrInvalidWhence
	}
	if abs < 0 {
		return 0, ErrNegativePosition
	}
	b.off = abs
	return abs, nil
}

var _ io.Writer = (*Bucket)(nil)

// Write implements the io.Writer interface.
// Write appends the contents of p to last read or wirte position of bucket. Growing the bucket as needed.
// Cursor shiftes if allowed from access modes.
// More details in method WriteAt.
func (b *Bucket) Write(p []byte) (n int, err error) {
	off := b.off
	if b.isAppend {
		// append guaranteed into end of the file
		off = int64(len(b.data))
	}
	return b.WriteAt(p, off)
}

var _ io.WriterAt = (*Bucket)(nil)

// WriteAt implements the io.WriterAt interface.
// WriteAt append the contents of p to a specified position of bucket. Growing the bucket as needed.
// Cursor shiftes if allowed from access modes.
//
// https://www.gnu.org/software/libc/manual/html_node/Operating-Modes.html
// The bit that enables append mode for the file. If set, then all write operations write the data at the end of the file, extending it, regardless of the current file position. This is the only reliable way to append to a file. In append mode, you are guaranteed that the data you write will always go to the current end of the file, regardless of other processes writing to the file. Conversely, if you simply set the file position to the end of file and write, then another process can extend the file after you set the file position but before you write, resulting in your data appearing someplace before the real end of file.
func (b *Bucket) WriteAt(p []byte, pos int64) (n int, err error) {
	if !b.isAppend {
		b.off = pos
	}
	n, err = b.writeAt(p, pos)
	if err != nil {
		return 0, err
	}

	if !b.isAppend {
		b.off += int64(n)
	}
	return
}

// this code is coped fragment from aws WrtieAtBuffer.
// auto growing if needed.
func (b *Bucket) writeAt(p []byte, pos int64) (n int, err error) {
	pLen := len(p)
	expLen := pos + int64(pLen)
	if int64(len(b.data)) < expLen {
		if int64(cap(b.data)) < expLen {
			newBuf := make([]byte, expLen, int64(float64(expLen)))
			copy(newBuf, b.data)
			b.data = newBuf
		}
		b.data = b.data[:expLen]
	}
	copy(b.data[pos:], p)
	return pLen, nil
}

// Size returns the original length of the data.
func (b *Bucket) Size() int64 {
	return int64(len(b.data))
}

// Cap returns the value of capacity of the slice stored of the data.
func (b *Bucket) Cap() int64 {
	return int64(cap(b.data))
}

var (
	// ErrNegativePosition means invalid arguments for Seek method.
	ErrNegativePosition = errors.New("negative position")
	// ErrNegativeSize means invalid arguments for Truncate method.
	ErrNegativeSize = errors.New("negative size")
	// ErrInvalidWhence means invalid arguments for Seek method.
	ErrInvalidWhence = errors.New("invalid whence")
)
