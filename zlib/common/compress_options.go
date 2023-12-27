package common

import "github.com/MeenaAlfons/go-zlib/zlib/capi"

type HeaderType int

const (
	HeaderTypeZlib HeaderType = iota
	HeaderTypeRaw
	// HeaderTypeGzip not supported yet.
)

type StrategyType int

const (
	StrategyDefault     StrategyType = StrategyType(capi.Z_DEFAULT_STRATEGY)
	StrategyFiltered    StrategyType = StrategyType(capi.Z_FILTERED)
	StrategyHuffmanOnly StrategyType = StrategyType(capi.Z_HUFFMAN_ONLY)
	StrategyRLE         StrategyType = StrategyType(capi.Z_RLE)
	StrategyFixed       StrategyType = StrategyType(capi.Z_FIXED)
)

func DefaultCompressOptions() CompressOptions {
	return &compressOptions{
		level:       2,
		windowBits:  15,
		header:      HeaderTypeZlib,
		memoryLevel: 2,
		strategy:    StrategyDefault,
		bufferSize:  1024,
	}
}

type CompressOptions interface {
	Level() int
	WindowBits() int
	Header() HeaderType
	MemoryLevel() int
	Strategy() StrategyType
	BufferSize() int
	InitialDictionary() []byte

	WithLevel(level int) CompressOptions
	WithWindowBits(windowBits int) CompressOptions
	WithHeader(header HeaderType) CompressOptions
	WithMemoryLevel(memoryLevel int) CompressOptions
	WithStrategy(strategy StrategyType) CompressOptions
	WithBufferSize(bufferSize int) CompressOptions
	WithInitialDictionary(initialDictionary []byte) CompressOptions
}

type compressOptions struct {
	level             int
	windowBits        int
	header            HeaderType
	memoryLevel       int
	strategy          StrategyType
	initialDictionary []byte

	bufferSize int
}

func (opts *compressOptions) Level() int {
	return opts.level
}

func (opts *compressOptions) WindowBits() int {
	return opts.windowBits
}

func (opts *compressOptions) Header() HeaderType {
	return opts.header
}

func (opts *compressOptions) MemoryLevel() int {
	return opts.memoryLevel
}

func (opts *compressOptions) Strategy() StrategyType {
	return opts.strategy
}

func (opts *compressOptions) BufferSize() int {
	return opts.bufferSize
}

func (opts *compressOptions) InitialDictionary() []byte {
	return opts.initialDictionary
}

func (opts *compressOptions) WithLevel(level int) CompressOptions {
	opts.level = level
	return opts
}

func (opts *compressOptions) WithWindowBits(windowBits int) CompressOptions {
	opts.windowBits = windowBits
	return opts
}

func (opts *compressOptions) WithHeader(header HeaderType) CompressOptions {
	opts.header = header
	return opts
}

func (opts *compressOptions) WithMemoryLevel(memoryLevel int) CompressOptions {
	opts.memoryLevel = memoryLevel
	return opts
}

func (opts *compressOptions) WithStrategy(strategy StrategyType) CompressOptions {
	opts.strategy = strategy
	return opts
}

func (opts *compressOptions) WithBufferSize(bufferSize int) CompressOptions {
	opts.bufferSize = bufferSize
	return opts
}

func (opts *compressOptions) WithInitialDictionary(initialDictionary []byte) CompressOptions {
	opts.initialDictionary = initialDictionary
	return opts
}
