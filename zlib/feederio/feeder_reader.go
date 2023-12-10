package feederio

import (
	"io"

	"github.com/MeenaAlfons/go-zlib/zlib/compression"
	"github.com/MeenaAlfons/go-zlib/zlib/utils"
)

func NewFeederReader(reader io.Reader, feeder compression.FeederConsumer, bufferSize int) io.ReadCloser {
	// When the stream arrives at the end of a buffer, its internal state would refer to a position past the end of the buffer.
	// This results in error: "found pointer to free object".
	// To avoid this case, we reserve one byte at the end of the buffer so that the final state will not point past the end of the buffer.
	// The last byte of the input buffer is reserved for memory safety reasons.
	// Note that the output buffer is protected by FeederConsumerSafeOutputBuffer.
	inputBuffer := make([]byte, bufferSize)
	inputBuffer = inputBuffer[:len(inputBuffer)-1]

	return &feederReader{
		reader:       reader,
		feeder:       feeder,
		zInputBuffer: inputBuffer,
	}
}

type feederReader struct {
	feeder compression.FeederConsumer
	reader io.Reader

	zInputBuffer []byte
}

func (r *feederReader) Read(p []byte) (n int, err error) {
	if r.feeder.CanCallConsume() {
		n, err = r.feeder.Consume(p)
		return
	}

	n, err = r.reader.Read(r.zInputBuffer)
	utils.Debug("FeederReader.Read after reader.Read n:%d, err:%v", n, err)
	if err != nil && err != io.EOF {
		return n, err
	}

	flush := compression.NoFlush
	if err == io.EOF {
		flush = compression.Finish
	}
	if flush == compression.Finish || n > 0 {
		// Only feed data with length > 0 or flush == true
		utils.Debug("FeederReader.Read Calling feeder.Feed input: %p, n: %d, flush: %v", &r.zInputBuffer[0], n, flush)
		return r.feeder.Feed(r.zInputBuffer[:n], flush, p)
	}
	return 0, nil
}

func (r *feederReader) Close() error {
	return nil
}
