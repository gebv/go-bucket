package bucket

import "os"

type opts struct {
	// if true the writer to append guaranteed into end of the file
	isAppend bool
}

type opt func(o *opts)

// SetAppend sets is append flag via bool value.
func SetAppend(v bool) opt {
	return func(o *opts) {
		o.isAppend = v
	}
}

// SetOpModes sets is append flag via bits of operation modes.
func SetOpModes(flag int) opt {
	return func(o *opts) {
		o.isAppend = flag&os.O_APPEND != 0
	}
}
