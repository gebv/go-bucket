package bucket

import (
	"errors"
	"io"
	"os"
)

// NewBucket returns bucket with initial data and sets file access modes.
// Implements io.ReaderAt, io.WriterAt, io.Seeker, io.Closer interfaces.
// Shared cursor for writer and reader.
func NewBucket(dat []byte, mode int) *bucket {
	return &bucket{
		data:  dat,
		modes: mode,
	}
}

// bucket this is special container store data and implements io.ReaderAt, io.WriterAt, io.Seeker, io.Closer interfaces.
type bucket struct {
	// access modes of content (from linked file of outside)
	modes int
	// bucket contents
	data []byte
	// true if was called Close function
	closed bool
	// true if was modified data
	changed bool
	// position of cursor for reader and writer
	off int64
}

// Reset resets the bucket to be empty, but capacity is not changes.
// Nothing happens if
// - is closed
// - access modes has os.O_RDONLY
func (b *bucket) Reset() {
	if b.closed == true {
		return
	}
	if isReadOnly(b.modes) {
		return
	}
	b.data = b.data[:0]
	b.off = 0
	b.changed = true
}

// Truncate truncates container size data to specifed size. Growing if the size exceeds the current data size.
// Continues to use the same allocated storage.
// Returns errors if is closed or if there is no access for the operation.
func (b *bucket) Truncate(size int64) error {
	if b.closed == true {
		return os.ErrClosed
	}
	if isReadOnly(b.modes) {
		return ErrNoAccess
	}
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
	b.changed = true
	return nil
}

var _ io.Reader = (*bucket)(nil)

// Read implements the io.Reader interface.
// Cursor is shiftes. Otherwise, the behavior is the same as ReadAt.
func (b *bucket) Read(buf []byte) (n int, err error) {
	n, err = b.ReadAt(buf, b.off)
	if err != nil {
		return 0, err
	}
	b.off += int64(n)
	return n, nil
}

var _ io.ReaderAt = (*bucket)(nil)

// Read implements the io.ReaderAt interface.
// Returns error if is closed or if there is no access for the operation or standard error for the reader.
// Does not change the internal state.
func (b *bucket) ReadAt(buf []byte, off int64) (n int, err error) {
	if b.closed == true {
		return 0, os.ErrClosed
	}
	if isWriteOnly(b.modes) {
		return 0, ErrNoAccess
	}
	if len(buf) > 0 && off == int64(len(b.data)) {
		return 0, io.EOF
	}
	if off > int64(len(b.data)) {
		return 0, io.ErrUnexpectedEOF
	}
	n = copy(buf, b.data[off:])
	return
}

var _ io.Seeker = (*bucket)(nil)

// Seek implements the io.Seeker interface.
func (b *bucket) Seek(offset int64, whence int) (int64, error) {
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

var _ io.Writer = (*bucket)(nil)

// Write implements the io.Writer interface.
// Write appends the contents of p to last read or wirte position of bucket. Growing the bucket as needed.
// Cursor shiftes if allowed from access modes.
// More details in method WriteAt.
func (b *bucket) Write(p []byte) (n int, err error) {
	if b.closed == true {
		return 0, os.ErrClosed
	}
	if isReadOnly(b.modes) {
		return 0, ErrNoAccess
	}

	off := b.off
	if isAppend(b.modes) {
		// append guaranteed into end of the file
		off = int64(len(b.data))
	}
	return b.WriteAt(p, off)
}

var _ io.WriterAt = (*bucket)(nil)

// WriteAt implements the io.WriterAt interface.
// WriteAt append the contents of p to a specified position of bucket. Growing the bucket as needed.
// Cursor shiftes if allowed from access modes.
//
// https://www.gnu.org/software/libc/manual/html_node/Operating-Modes.html
// The bit that enables append mode for the file. If set, then all write operations write the data at the end of the file, extending it, regardless of the current file position. This is the only reliable way to append to a file. In append mode, you are guaranteed that the data you write will always go to the current end of the file, regardless of other processes writing to the file. Conversely, if you simply set the file position to the end of file and write, then another process can extend the file after you set the file position but before you write, resulting in your data appearing someplace before the real end of file.
func (b *bucket) WriteAt(p []byte, pos int64) (n int, err error) {
	if b.closed == true {
		return 0, os.ErrClosed
	}
	if isReadOnly(b.modes) {
		return 0, ErrNoAccess
	}

	if !isAppend(b.modes) {
		b.off = pos
	}
	n, err = b.writeAt(p, pos)
	if err != nil {
		return 0, err
	}
	b.changed = true
	if !isAppend(b.modes) {
		b.off += int64(n)
	}
	return
}

// this code is coped fragment from aws WrtieAtBuffer.
func (b *bucket) writeAt(p []byte, pos int64) (n int, err error) {
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

var _ io.Closer = (*bucket)(nil)

// Close implements the io.Closer interface.
func (b *bucket) Close() error {

	if b.closed {
		return os.ErrClosed
	}
	b.closed = true
	return nil
}

// Size returns the original length of the data.
func (b *bucket) Size() int64 {
	return int64(len(b.data))
}

// Size returns the value of capacity of the slice stored of the data.
func (b *bucket) Cap() int64 {
	return int64(cap(b.data))
}

// Changed returns true if data has been changed.
func (b *bucket) Changed() bool {
	return b.changed
}

// Closed returns true if has been called Close.
func (b *bucket) Closed() bool {
	return b.closed
}

// Mode returns allowed i/o operating modes.
func (b *bucket) Mode() int {
	return b.modes
}

var (
	ErrNegativePosition = errors.New("negative position")
	ErrNegativeSize     = errors.New("negative size")
	ErrInvalidWhence    = errors.New("invalid whence")
	// ErrNoAccess means that an operation rejected by a i/o operating modes.
	ErrNoAccess = errors.New("no access")
)
