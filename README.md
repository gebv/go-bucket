# go-bucket

Bucket is a variable-sized buffer of bytes with Read and Write methods. More like to `bytes.Buffer`. But includes:

- marker indicating that the data has been changed
- includes and apply flags of i/o operating modes

Also implement `io.ReaderAt`, `io.WriterAt`, `io.Seeker`, `io.Closer`.

