package zlib

import (
	"io"

	"github.com/MeenaAlfons/go-zlib/zlib/common"
	"github.com/MeenaAlfons/go-zlib/zlib/compression"
	"github.com/MeenaAlfons/go-zlib/zlib/feederio"
)

// NewCompressReader reads uncompressed data from target and compresses it.
// It returns a ReadCloser that reads compressed data.
func NewCompressReader(target io.Reader, opts common.CompressOptions) (io.ReadCloser, error) {
	zcompressor, err := compression.NewCompressor(opts)
	if err != nil {
		return nil, err
	}

	r := &compressReader{
		impl: feederio.NewFeederReader(target, zcompressor, opts.BufferSize()),
	}
	return r, nil
}

type compressReader struct {
	impl io.ReadCloser
}

// Read reads compressed data resulting from compressing the data read from target.
func (r *compressReader) Read(p []byte) (n int, err error) {
	return r.impl.Read(p)
}

// Close method is unnecessary at this moment.
func (r *compressReader) Close() error {
	return r.impl.Close()
}
