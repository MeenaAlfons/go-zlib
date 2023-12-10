package feederio

import (
	"fmt"
	"io"

	"github.com/MeenaAlfons/go-zlib/zlib/common"
	"github.com/MeenaAlfons/go-zlib/zlib/compression"
	"github.com/MeenaAlfons/go-zlib/zlib/utils"
)

func NewFeederWriter(writer io.Writer, feeder compression.FeederConsumer, bufferSize int) common.WriteFlushCloser {
	// When the stream arrives at the end of a buffer, its internal state would refer to a position past the end of the buffer.
	// This results in error: "found pointer to free object".
	// To avoid this case, we reserve one byte at the end of the buffer so that the final state will not point past the end of the buffer.
	// The last byte of the input buffer is reserved for memory safety reasons.
	// Note that the output buffer is protected by FeederConsumerSafeOutputBuffer.
	inputBuffer := make([]byte, bufferSize)
	inputBuffer = inputBuffer[:len(inputBuffer)-1]

	return &feederWriter{
		feeder:        feeder,
		writer:        writer,
		zInputBuffer:  inputBuffer,
		zOutputBuffer: make([]byte, bufferSize),
	}
}

type feederWriter struct {
	feeder compression.FeederConsumer
	writer io.Writer

	zInputBuffer  []byte
	zOutputBuffer []byte
}

func (r *feederWriter) writeSome(p []byte) (int, error) {
	newP := p
	// newP := make([]byte, len(p))
	// copy(newP, p)

	// we can get how much of the input was consumed by the zlib
	// and return it as the number of bytes written

	flush := compression.NoFlush
	utils.Debug("FeederWriter.Read Calling feeder.Feed input: %p, n: %d, flush: %v", &newP[0], len(newP), flush)
	n1, err1 := r.feeder.Feed(newP, flush, r.zOutputBuffer)
	if err1 != nil && err1 != io.EOF {
		return len(p), err1
	}

	n2, err2 := r.writer.Write(r.zOutputBuffer[:n1])
	if n1 != n2 {
		return len(p), fmt.Errorf("short write %w", io.ErrShortWrite)
	}
	if err2 != nil {
		return len(p), err2
	}

	for r.feeder.CanCallConsume() {
		n1, err1 = r.feeder.Consume(r.zOutputBuffer)
		if err1 != nil && err1 != io.EOF {
			return len(p), err1
		}

		n2, err2 := r.writer.Write(r.zOutputBuffer[:n1])
		if n1 != n2 {
			return len(p), fmt.Errorf("short write %w", io.ErrShortWrite)
		}
		if err2 != nil {
			return len(p), err2
		}
	}

	// Return err1 because it could be io.EOF
	return len(p), err1
}

func (r *feederWriter) Write(p []byte) (int, error) {
	// we can get how much of the input was consumed by the zlib
	// and return it as the number of bytes written

	// Copy p into zInputBuffer because we need to keep the data around
	// until the zlib consumes it. Note that p could be reused or freed by the caller.
	// But it deosn't make sense to end this method with the input not fully consumed!!!
	// TODO: I think we need to reset the input so that we don't keep a pointer to it.
	inputIndex := 0
	for inputIndex < len(p) {
		n := copy(r.zInputBuffer, p[inputIndex:])
		inputIndex += n
		_, err := r.writeSome(r.zInputBuffer[:n])
		if err != nil {
			return inputIndex, err
		}
	}
	return inputIndex, nil
}

func (r *feederWriter) Flush() error {
	if isDone, reason := r.feeder.IsDoneWithReason(); isDone {
		if reason != nil {
			return reason
		}
		// return io.EOF
		return nil
	}

	flush := compression.SyncFlush
	n1, err1 := r.feeder.Feed(nil, flush, r.zOutputBuffer)
	if err1 != nil && err1 != io.EOF {
		return err1
	}

	n2, err2 := r.writer.Write(r.zOutputBuffer[:n1])
	if n1 != n2 {
		return fmt.Errorf("short write %w", io.ErrShortWrite)
	}
	if err2 != nil {
		return err2
	}

	for r.feeder.CanCallConsume() {
		n1, err1 = r.feeder.Consume(r.zOutputBuffer)
		if err1 != nil && err1 != io.EOF {
			return err1
		}

		n2, err2 := r.writer.Write(r.zOutputBuffer[:n1])
		if n1 != n2 {
			return fmt.Errorf("short write %w", io.ErrShortWrite)
		}
		if err2 != nil {
			return err2
		}
	}

	// This is wrong. Flush only pushes the current state out. It doesn't end the stream.
	// if err1 != io.EOF {
	// 	return fmt.Errorf("err1 must be io.EOF")
	// }

	return nil
}

func (r *feederWriter) Close() error {
	if isDone, reason := r.feeder.IsDoneWithReason(); isDone {
		if reason != nil {
			return reason
		}
		// return io.EOF
		return nil
	}

	flush := compression.Finish
	n1, err1 := r.feeder.Feed(nil, flush, r.zOutputBuffer)
	if err1 != nil && err1 != io.EOF {
		return err1
	}

	n2, err2 := r.writer.Write(r.zOutputBuffer[:n1])
	if n1 != n2 {
		return fmt.Errorf("short write %w", io.ErrShortWrite)
	}
	if err2 != nil {
		return err2
	}

	for r.feeder.CanCallConsume() {
		n1, err1 = r.feeder.Consume(r.zOutputBuffer)
		if err1 != nil && err1 != io.EOF {
			return err1
		}

		n2, err2 := r.writer.Write(r.zOutputBuffer[:n1])
		if n1 != n2 {
			return fmt.Errorf("short write %w", io.ErrShortWrite)
		}
		if err2 != nil {
			return err2
		}
	}

	// This is wrong. Flush only pushes the current state out. It doesn't end the stream.
	if err1 != io.EOF {
		return fmt.Errorf("err1 must be io.EOF. It was %w", err1)
	}

	return nil
}
