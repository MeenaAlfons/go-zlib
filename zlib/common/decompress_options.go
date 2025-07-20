package common

func DefaultDecompressOptions() DecompressOptions {
	return &decompressOptions{
		windowBits:  15,
		header:      HeaderTypeZlib,
		bufferSize:  1024,
		autoDestroy: true,
	}
}

type DecompressOptions interface {
	WindowBits() int
	Header() HeaderType
	BufferSize() int
	InitialDictionary() []byte
	AutoDestroy() bool

	WithWindowBits(windowBits int) DecompressOptions
	WithHeader(header HeaderType) DecompressOptions
	WithBufferSize(bufferSize int) DecompressOptions
	WithInitialDictionary(initialDictionary []byte) DecompressOptions
	WithAutoDestroy(autoDestroy bool) DecompressOptions
}

type decompressOptions struct {
	windowBits        int
	header            HeaderType
	initialDictionary []byte

	bufferSize  int
	autoDestroy bool
}

func (opts *decompressOptions) WindowBits() int {
	return opts.windowBits
}

func (opts *decompressOptions) Header() HeaderType {
	return opts.header
}

func (opts *decompressOptions) BufferSize() int {
	return opts.bufferSize
}

func (opts *decompressOptions) InitialDictionary() []byte {
	return opts.initialDictionary
}

func (opts *decompressOptions) AutoDestroy() bool {
	return opts.autoDestroy
}

func (opts *decompressOptions) WithWindowBits(windowBits int) DecompressOptions {
	opts.windowBits = windowBits
	return opts
}

func (opts *decompressOptions) WithHeader(header HeaderType) DecompressOptions {
	opts.header = header
	return opts
}

func (opts *decompressOptions) WithBufferSize(bufferSize int) DecompressOptions {
	opts.bufferSize = bufferSize
	return opts
}

func (opts *decompressOptions) WithInitialDictionary(initialDictionary []byte) DecompressOptions {
	opts.initialDictionary = initialDictionary
	return opts
}

func (opts *decompressOptions) WithAutoDestroy(autoDestroy bool) DecompressOptions {
	opts.autoDestroy = autoDestroy
	return opts
}
