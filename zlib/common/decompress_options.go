package common

func DefaultDecompressOptions() DecompressOptions {
	return &decompressOptions{
		windowBits: 15,
		header:     HeaderTypeZlib,
		bufferSize: 1024,
	}
}

type DecompressOptions interface {
	WindowBits() int
	Header() HeaderType
	BufferSize() int

	WithWindowBits(windowBits int) DecompressOptions
	WithHeader(header HeaderType) DecompressOptions
	WithBufferSize(bufferSize int) DecompressOptions
}

type decompressOptions struct {
	windowBits int
	header     HeaderType

	bufferSize int
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
