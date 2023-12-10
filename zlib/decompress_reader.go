package zlib

import (
	"io"

	"github.com/MeenaAlfons/go-zlib/zlib/common"
	"github.com/MeenaAlfons/go-zlib/zlib/compression"
	"github.com/MeenaAlfons/go-zlib/zlib/feederio"
)

// NewDecompressReader reads compressed data from target and decompresses it.
// It returns a ReadCloser that reads decompressed data.
func NewDecompressReader(target io.Reader, opts common.DecompressOptions) (io.ReadCloser, error) {
	zcompressor, err := compression.NewDecompressor(opts)
	if err != nil {
		return nil, err
	}

	r := &decompressReader{
		impl: feederio.NewFeederReader(target, zcompressor, opts.BufferSize()),
	}
	return r, nil
}

type decompressReader struct {
	impl io.ReadCloser
}

// Read reads decompressed data resulting from decompressing the data read from target.
func (r *decompressReader) Read(p []byte) (n int, err error) {
	return r.impl.Read(p)
}

// Close method is unnecessary at this moment.
func (r *decompressReader) Close() error {
	return r.impl.Close()
}
