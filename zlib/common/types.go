package common

import "io"

type WriteFlushCloser interface {
	Flush() error
	io.Writer
	io.Closer
}

type WriteFlushCloseResetter interface {
	WriteFlushCloser
	Reset(writer io.Writer) error
}

type ReadCloseResetter interface {
	io.ReadCloser
	Reset(reader io.Reader) error
}
