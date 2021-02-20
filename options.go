package bucket

import "os"

type bucketOpts struct {
	// if true the writer to append guaranteed into end of the file
	isAppend bool
}

type optBucket func(o *bucketOpts)

// SetAppend sets is append flag via bool value.
func SetAppend(v bool) optBucket {
	return func(o *bucketOpts) {
		o.isAppend = v
	}
}

// SetOpModes sets is append flag via bits of operation modes.
func SetOpModes(flag int) optBucket {
	return func(o *bucketOpts) {
		o.isAppend = flag&os.O_APPEND != 0
	}
}

type optCloser func(o *closerOpts)

// CloseHook sets callback function for close method.
func CloseHook(fn func(changed bool, i BucketIface) error) optCloser {
	return func(o *closerOpts) {
		o.closeFn = fn
	}
}

type closerOpts struct {
	// callback function of close function
	// will be called once the first time the io.Closer interface is closed. Passing interface  will not be closed yet.
	closeFn func(changed bool, i BucketIface) error
}
