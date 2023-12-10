package zlib

import (
	"io"

	"github.com/MeenaAlfons/go-zlib/zlib/common"
	"github.com/MeenaAlfons/go-zlib/zlib/compression"
	"github.com/MeenaAlfons/go-zlib/zlib/feederio"
)

// NewDecompressWriter writes decompressed data to target.
// It returns a WriteFlushCloser which is used to write compressed data to be decompressed.
func NewDecompressWriter(target io.Writer, opts common.DecompressOptions) (common.WriteFlushCloser, error) {
	zcompressor, err := compression.NewDecompressor(opts)
	if err != nil {
		return nil, err
	}

	r := &decompressWriter{
		impl: feederio.NewFeederWriter(target, zcompressor, opts.BufferSize()),
	}
	return r, nil
}

type decompressWriter struct {
	impl common.WriteFlushCloser
}

// Write writes compressed data which will be decompressed and written to target.
func (w *decompressWriter) Write(p []byte) (n int, err error) {
	return w.impl.Write(p)
}

// Flush flushes the decompressed data so far to target.
func (w *decompressWriter) Flush() error {
	return w.impl.Flush()
}

// Close method concludes the decompression process and flushes the remaining decompressed data to target.
// Write and Flush methods can not be called after Close.
func (w *decompressWriter) Close() error {
	return w.impl.Close()
}
