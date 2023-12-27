package compression

import (
	"fmt"
	"io"

	"github.com/MeenaAlfons/go-zlib/zlib/capi"
	"github.com/MeenaAlfons/go-zlib/zlib/common"
	"github.com/MeenaAlfons/go-zlib/zlib/utils"
)

// NewCompressor creates a new compressor FeederConsumer with the given options.
func NewCompressor(opts common.CompressOptions) (FeederConsumer, error) {
	c := &compressor{
		zstream: capi.NewZStream(),
	}

	ret := c.zstream.DeflateInit2(opts.Level(), zWindowBits(opts), opts.MemoryLevel(), int(opts.Strategy()))
	if ret != capi.Z_OK {
		return nil, capi.ZError(ret)
	}

	if opts.InitialDictionary() != nil {
		ret = c.zstream.DeflateSetDictionary(opts.InitialDictionary())
		if ret != capi.Z_OK {
			return nil, capi.ZError(ret)
		}
	}

	return newFeederConsumerSafeOutputBuffer(c), nil
}

type compressor struct {
	zstream capi.ZStream

	lastFlush     Flush
	hasMoreOutput bool

	// StreamEnd is called when the stream has successfully ended or when an unrecoverable error has occurred
	streamEndHasBeenCalled bool

	// This is the last error returned when the streamEnd was called.
	// It could be io.EOF if everything went well or an error otherwise.
	streamEndError error

	// This is the reason that was passed to endStream
	// It is stored separately from streamEndError because streamEndError may include the error from deflateEnd even if streamEndReason is nil
	streamEndReason error
}

// Make sure that the input buffer has capacity larger than its size by at least one.
// This is to avoid the case where the stream ends at the end of the buffer which would
// result in an internal state that points past the end of the buffer and causes an error
// "found pointer to free object".
func (c *compressor) Feed(input []byte, flush Flush, outputBuffer []byte) (int, error) {
	if c.lastFlush == Finish {
		return 0, fmt.Errorf("zlib: cannot call Feed after it has been called with flush = Finish. Call Consume instead")
	}

	if c.streamEndHasBeenCalled {
		// This only happens when the stream has ended because of an error. Otherwise, c.lastFlush would be Finish and we'll not get to this condition.
		// A previous call to Feed or Consume would have already returned the error. The caller should recognize the error and stop feeding more data.
		return 0, fmt.Errorf("zlib: stream has ended and cannot be used anymore. Stream ended with reason: %v, err: %v", c.streamEndReason, c.streamEndError)
	}

	if c.CanCallConsume() {
		return 0, fmt.Errorf("zlib: cannot call Feed when there is still output to be consumed. Call Consume instead. Always check CanCallConsume")
	}

	c.lastFlush = flush
	zflush := zFlush(c.lastFlush)
	c.zstream.SetInput(input)
	c.zstream.SetOutput(outputBuffer)
	ret := c.zstream.Deflate(zflush)
	have := c.zstream.ProducedOutput()
	c.hasMoreOutput = c.zstream.OutputBufferIsFull()
	err := c.processReturnValue(ret)
	utils.Debug("ZCompress Feed flush:%v have:%v err:%v hasMoreOutput:%v len(input):%v len(output):%v, output:%x", c.lastFlush, have, err, c.hasMoreOutput, len(input), len(outputBuffer), outputBuffer[:have])
	return have, err
}

func (c *compressor) IsDoneWithReason() (bool, error) {
	return c.streamEndHasBeenCalled, c.streamEndReason
}

func (c *compressor) CanCallConsume() bool {
	return !c.streamEndHasBeenCalled && c.hasMoreOutput
}

func (c *compressor) Consume(outputBuffer []byte) (int, error) {
	if c.streamEndHasBeenCalled {
		// This happens when the stream has ended
		// - successfully
		// - or because of an error
		// The caller should recognize both cases from previous calls to Feed or Consume.
		// However, we return the streamEndError again anyway to make calling Consume idempotent.
		return 0, c.streamEndError
	}

	if !c.hasMoreOutput {
		// We may return something here to indicate that there is no more output
		// To inform the caller that there is no point of calling Consume again
		// and it should call Feed instead.
		// We currently depend on the caller to check CanCallConsume
		// TIE_1: tie this case to a similar one in processReturnValue
		return 0, nil
	}

	zflush := zFlush(c.lastFlush)
	c.zstream.SetOutput(outputBuffer)
	ret := c.zstream.Deflate(zflush)
	have := c.zstream.ProducedOutput()
	c.hasMoreOutput = c.zstream.OutputBufferIsFull()
	err := c.processReturnValue(ret)
	utils.Debug("ZCompress Consume flush:%v have:%v err:%v hasMoreOutput:%v", c.lastFlush, have, err, c.hasMoreOutput)
	return have, err
}

func (c *compressor) processReturnValue(ret capi.ZConstant) error {
	if ret == capi.Z_STREAM_ERROR {
		// Z_STREAM_ERROR indicates that the stream state was inconsistent
		// which may happen if the stream was not initialized
		// Or the appplication is broken and altered the memory of the stream state.
		reason := fmt.Errorf("zlib: deflate failed with err: %w", capi.ZError(ret))
		return c.endStream(reason)
	}

	if !c.hasMoreOutput {
		// Output buffer is not full, indicating that no more output needs to be consumed
		// At this point, the input should have been fully consumed.
		if c.zstream.AvailIn() > 0 {
			// This should never happen, but who knows!
			reason := fmt.Errorf("zlib: no more output but the input is not fully consumed: %w", capi.ZError(ret))
			return c.endStream(reason)
		}

		if c.lastFlush == Finish {
			// If flush is Z_FINISH, then we should have consumed all input and output all data
			// and ret should be Z_STREAM_END
			if ret != capi.Z_STREAM_END {
				reason := fmt.Errorf("zlib: no more output, flush is requested, but the result is not Z_STREAM_END: %w", capi.ZError(ret))
				return c.endStream(reason)
			}

			err := c.endStream(nil)
			if err != nil {
				return err
			}
			return io.EOF
		}

		// We may return something here to indicate that there is no more output
		// To inform the caller that there is no point of calling Consume again
		// and it should call Feed instead.
		// We currently depend on the caller to check CanCallConsume
		// TIE_1: tie this case to a similar one in Consume
		return nil
	}

	// The output buffer is full. This could mean two things:
	// - There is more output to be consumed which is indicated by ret = Z_BUF_ERROR
	// - There is no more output to be consumed which is indicated by ret = Z_OK
	// We return nil in both cases.
	return nil
}

// deflateEnd needs to be called when ZStream is no longer needed
func (c *compressor) endStream(reason error) error {
	endRet := c.zstream.DeflateEnd()
	c.streamEndHasBeenCalled = true

	c.streamEndReason = reason
	c.streamEndError = processStreamEndError(reason, endRet)
	return c.streamEndError
}

func processStreamEndError(reason error, endRet capi.ZConstant) error {
	if reason != nil {
		return wrapWithDistructionNote(reason, capi.ZError(endRet))
	}

	if endRet != capi.Z_OK {
		return fmt.Errorf("zlib: the stream ended successfully but distruction failed: %w", capi.ZError(endRet))
	}

	return nil
}

func zFlush(flush Flush) capi.ZConstant {
	switch flush {
	case NoFlush:
		return capi.Z_NO_FLUSH
	case SyncFlush:
		return capi.Z_SYNC_FLUSH
	case Finish:
		return capi.Z_FINISH
	default:
		panic("invalid flush")
	}
}

func wrapWithDistructionNote(originalError error, destructionError error) error {
	return fmt.Errorf("%w. The stream is no longer usable and was destructed. The result of stream destruction is: %s", originalError, destructionError)
}

// TODO
// With flush set to Z_FINISH, this final set of deflate() calls will complete
// the output stream. Once that is done, subsequent calls of deflate() would
// return Z_STREAM_ERROR if the flush parameter is not Z_FINISH, and do no more
// processing until the state is reinitialized (call deflateInit).
// func (c *ZCompressor2) Reset() err {

// 	ret := c.zstream.DeflateInit2(opts.Level, opts.ZWindowBits(), opts.MemoryLevel, int(opts.Strategy))
// 	if ret != Z_OK {
// 		return nil, ZError(ret)
// 	}
// }
