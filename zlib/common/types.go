package common

import "io"

type WriteFlushCloser interface {
	Flush() error
	io.Writer
	io.Closer
}
