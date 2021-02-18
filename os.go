package bucket

import "os"

func isCreate(flag int) bool {
	return flag&os.O_CREATE != 0
}

func isAppend(flag int) bool {
	return flag&os.O_APPEND != 0
}

func isTruncate(flag int) bool {
	return flag&os.O_TRUNC != 0
}

func isReadAndWrite(flag int) bool {
	return flag&os.O_RDWR != 0
}

func isReadOnly(flag int) bool {
	return flag == os.O_RDONLY
}

func isWriteOnly(flag int) bool {
	return flag&os.O_WRONLY != 0
}

func isSymlink(m os.FileMode) bool {
	return m&os.ModeSymlink != 0
}

func fileAccessModesPretty(flag int) string {
	flags := "File access modes: "
	if isCreate(flag) {
		flags += "create "
	}
	if isAppend(flag) {
		flags += "append "
	}
	if isTruncate(flag) {
		flags += "truncate "
	}
	if isReadAndWrite(flag) {
		flags += "read&&write "
	}
	if isReadOnly(flag) {
		flags += "readOnly "
	}
	if isWriteOnly(flag) {
		flags += "writeOnly "
	}
	return flags
}
