package test

import (
	"fmt"
	"io"

	"github.com/MeenaAlfons/go-zlib/zlib"
	"github.com/MeenaAlfons/go-zlib/zlib/common"
)

type ZLib interface {
	NewCompressReader(io.Reader, common.CompressOptions) (io.ReadCloser, error)
	NewCompressWriter(io.Writer, common.CompressOptions) (common.WriteFlushCloseResetter, error)
	NewDecompressReader(io.Reader, common.DecompressOptions) (io.ReadCloser, error)
	NewDecompressWriter(io.Writer, common.DecompressOptions) (common.WriteFlushCloseResetter, error)
}

func getZlibs() []ZLib {
	return []ZLib{
		&ZLibDefault{},
		NewZlibReset(),
	}
}

type ZLibDefault struct{}

func (z *ZLibDefault) NewCompressReader(r io.Reader, opts common.CompressOptions) (io.ReadCloser, error) {
	return zlib.NewCompressReader(r, opts)
}
func (z *ZLibDefault) NewDecompressReader(r io.Reader, opts common.DecompressOptions) (io.ReadCloser, error) {
	return zlib.NewDecompressReader(r, opts)
}
func (z *ZLibDefault) NewCompressWriter(w io.Writer, opts common.CompressOptions) (common.WriteFlushCloseResetter, error) {
	return zlib.NewCompressWriter(w, opts)
}

func (z *ZLibDefault) NewDecompressWriter(w io.Writer, opts common.DecompressOptions) (common.WriteFlushCloseResetter, error) {
	return zlib.NewDecompressWriter(w, opts)
}

type ZLibReset struct {
	compressWriter   map[common.CompressOptions]common.WriteFlushCloseResetter
	decompressWriter map[common.DecompressOptions]common.WriteFlushCloseResetter
}

func NewZlibReset() *ZLibReset {
	return &ZLibReset{
		compressWriter:   make(map[common.CompressOptions]common.WriteFlushCloseResetter),
		decompressWriter: make(map[common.DecompressOptions]common.WriteFlushCloseResetter),
	}
}

// Readers don't support Reset for now.
func (z *ZLibReset) NewCompressReader(r io.Reader, opts common.CompressOptions) (io.ReadCloser, error) {
	return zlib.NewCompressReader(r, opts)
}
func (z *ZLibReset) NewDecompressReader(r io.Reader, opts common.DecompressOptions) (io.ReadCloser, error) {
	return zlib.NewDecompressReader(r, opts)
}

func (z *ZLibReset) NewCompressWriter(w io.Writer, opts common.CompressOptions) (common.WriteFlushCloseResetter, error) {
	opts = opts.WithAutoDestroy(false)

	if v, ok := z.compressWriter[opts]; ok {
		fmt.Println("Reusing compressor for options:", opts)
		err := v.Reset(w)
		if err != nil {
			return nil, fmt.Errorf("failed to reset compressor: %w", err)
		}
		return v, nil
	}

	fmt.Println("Creating new compressor for options:", opts)
	compressor, err := zlib.NewCompressWriter(w, opts)
	if err != nil {
		return nil, err
	}
	z.compressWriter[opts] = compressor
	return compressor, nil
}

func (z *ZLibReset) NewDecompressWriter(w io.Writer, opts common.DecompressOptions) (common.WriteFlushCloseResetter, error) {
	opts = opts.WithAutoDestroy(false)

	if v, ok := z.decompressWriter[opts]; ok {
		fmt.Printf("Reusing decompressor %p for options: %v\n", v, opts)
		err := v.Reset(w)
		if err != nil {
			return nil, fmt.Errorf("failed to reset decompressor: %w", err)
		}
		return v, nil
	}

	fmt.Println("Creating new decompressor for options:", opts)

	decompressor, err := zlib.NewDecompressWriter(w, opts)
	if err != nil {
		return nil, err
	}

	z.decompressWriter[opts] = decompressor
	return decompressor, nil
}
