package zlib

import (
	"io"

	"github.com/MeenaAlfons/go-zlib/zlib/common"
	"github.com/MeenaAlfons/go-zlib/zlib/compression"
	"github.com/MeenaAlfons/go-zlib/zlib/feederio"
)

// NewCompressWriter writes compressed data to target.
// It returns a WriteFlushCloser which is used to write decompressed data to be compressed.
func NewCompressWriter(target io.Writer, opts common.CompressOptions) (common.WriteFlushCloseResetter, error) {
	zcompressor, err := compression.NewCompressor(opts)
	if err != nil {
		return nil, err
	}

	r := &compressWriter{
		impl: feederio.NewFeederWriter(target, zcompressor, opts.BufferSize()),
	}
	return r, nil
}

type compressWriter struct {
	impl common.WriteFlushCloseResetter
}

// Write writes decompressed data which will be compressed and written to target.
func (w *compressWriter) Write(p []byte) (n int, err error) {
	return w.impl.Write(p)
}

// Flush flushes the compressed data so far to target.
func (w *compressWriter) Flush() error {
	return w.impl.Flush()
}

// Close method concludes the compression process and flushes the remaining compressed data to target.
// Write and Flush methods can not be called after Close.
func (w *compressWriter) Close() error {
	return w.impl.Close()
}

func (w *compressWriter) Reset(target io.Writer) error {
	return w.impl.Reset(target)
}
