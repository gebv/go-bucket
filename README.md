# go-bucket

![CI Status](https://github.com/gebv/go-bucket/workflows/tests/badge.svg)
[![Go Report Card](https://goreportcard.com/badge/github.com/gebv/go-bucket)](https://goreportcard.com/report/github.com/gebv/go-bucket)
[![codecov](https://codecov.io/gh/gebv/go-bucket/branch/master/graph/badge.svg)](https://codecov.io/gh/gebv/go-bucket)


Bucket is a variable-sized buffer of bytes with Read and Write methods. More like to `bytes.Buffer`. But includes:

- implement `io.Reader`, `io.ReaderAt`, `io.Writer`, `io.WriterAt`, `io.Seeker`.
- optionally can be wrapped in `io.Closer` with detection of changes for callback function.
- the flag `os.O_APPEND` logic is being applied
<pre>
The bit that enables append mode for the file. If set, then all write operations write the data at the end of the file, extending it, regardless of the current file position. This is the only reliable way to append to a file. In append mode, you are guaranteed that the data you write will always go to the current end of the file, regardless of other processes writing to the file. Conversely, if you simply set the file position to the end of file and write, then another process can extend the file after you set the file position but before you write, resulting in your data appearing someplace before the real end of file.
</pre>
See more details of flags https://www.gnu.org/software/libc/manual/html_node/Operating-Modes.html


See example for [bucket](bucket_test.go#L11-43) and [closer](closer_test.go#L13-41).

TODO features
- (maybe) lazy load - init bucket with lazy loading data
- (experements) implement with io.Pipe for specific cases
