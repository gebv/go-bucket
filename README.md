# go-bucket

![CI Status](https://github.com/gebv/go-bucket/workflows/tests/badge.svg)
[![Go Report Card](https://goreportcard.com/badge/github.com/gebv/go-bucket)](https://goreportcard.com/report/github.com/gebv/go-bucket)
[![codecov](https://codecov.io/gh/gebv/go-bucket/branch/master/graph/badge.svg)](https://codecov.io/gh/gebv/go-bucket)


Bucket is a variable-sized buffer of bytes with Read and Write methods. More like to `bytes.Buffer`. But includes:

- marker indicating that the data has been changed
- includes and apply flags of i/o operating modes

Also implement `io.ReaderAt`, `io.WriterAt`, `io.Seeker`, `io.Closer`.

