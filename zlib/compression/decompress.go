package compression

import (
	"fmt"
	"io"

	"github.com/MeenaAlfons/go-zlib/zlib/capi"
	"github.com/MeenaAlfons/go-zlib/zlib/common"
	"github.com/MeenaAlfons/go-zlib/zlib/utils"
)

// NewDecompressor creates a new decompressor FeederConsumer with the given options.
func NewDecompressor(opts common.DecompressOptions) (FeederConsumer, error) {
	c := &decompressor{
		zstream: capi.NewZStream(),
	}

	ret := c.zstream.InflateInit2(zWindowBits(opts))
	if ret != capi.Z_OK {
		return nil, capi.ZError(ret)
	}

	return newFeederConsumerSafeOutputBuffer(c), nil
}

type decompressor struct {
	zstream capi.ZStream

	lastFlush     Flush
	hasMoreOutput bool

	// StreamEnd is called when the stream has successfully ended or when an unrecoverable error has occurred
	streamEndHasBeenCalled bool

	// This is the last error returned when the streamEnd was called.
	// It could be io.EOF if everything went well or an error otherwise.
	streamEndError error

	// This is the reason that was passed to endStream
	// It is stored separately from streamEndError because streamEndError may be include the error from inflateEnd
	streamEndReason error
}

// Make sure that the input buffer has capacity larger than its size by at least one.
// This is to avoid the case where the stream ends at the end of the buffer which would
// result in an internal state that points past the end of the buffer and causes an error
// "found pointer to free object".
func (c *decompressor) Feed(input []byte, flush Flush, outputBuffer []byte) (int, error) {
	if c.lastFlush == Finish {
		return 0, fmt.Errorf("feed: cannot call Feed after it has been called with flush = Finish. Call Consume instead")
	}

	if c.streamEndHasBeenCalled {
		// This happens in one of two cases:
		// - The stream has ended because of an error.
		//   A previous call to Feed or Consume would have already returned the error.
		//   The caller should recognize the error and stop feeding more data.
		// - the stream has ended because the decompression ends when the input data indicates an end block.
		//   A previous call to Consume or Feed should have already returned an io.EOF.
		//   The caller should recognize this case and stop feeding more data.
		return 0, fmt.Errorf("feed: stream has ended and cannot be used anymore. Stream ended with reason: %v, err: %v", c.streamEndReason, c.streamEndError)
	}

	if c.CanCallConsume() {
		return 0, fmt.Errorf("feed: cannot call Feed when there is still output to be consumed. Call Consume instead. Always check CanCallConsume")
	}

	c.lastFlush = flush
	zflush := zFlush(c.lastFlush)

	c.zstream.SetInput(input)
	c.zstream.SetOutput(outputBuffer)
	ret := c.zstream.Inflate(zflush)
	have := c.zstream.ProducedOutput()
	c.hasMoreOutput = c.zstream.OutputBufferIsFull()
	err := c.processReturnValue(ret)
	utils.Debug("ZDecompressor Feed flush:%v have:%v err:%v hasMoreOutput:%v len(input):%v len(output):%v", c.lastFlush, have, err, c.hasMoreOutput, len(input), len(outputBuffer))
	return have, err
}

func (c *decompressor) IsDoneWithReason() (bool, error) {
	return c.streamEndHasBeenCalled, c.streamEndReason
}

func (c *decompressor) CanCallConsume() bool {
	// This is not totally correct, I think there is a case where there is still more input but the decompression has ended, there is not more input needed, and the output buffer is not full.
	// In that case, there is no point of calling Consume again!
	// We probably have returned io.EOF already. but we need to store the fact that the decompression has ended so that we can return false here and probably return io.EOF for all future calls.
	// TODO: tie this case with the others. See processReturnValue.

	// As long as there is more input or more output, we can call Consume.
	// In case the decompression has ended while the input is not fully consumed, streamEndHasBeenCalled would be true and Consume doesn't need to be called again.
	// TODO: There was a case where more input is needed but I removed it. I think it is not needed.
	return !c.streamEndHasBeenCalled && c.hasMoreOutput
}

func (c *decompressor) Consume(outputBuffer []byte) (int, error) {
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
	ret := c.zstream.Inflate(zflush)
	have := c.zstream.ProducedOutput()
	c.hasMoreOutput = c.zstream.OutputBufferIsFull()
	err := c.processReturnValue(ret)
	utils.Debug("ZDecompressor Consume flush:%v have:%v err:%v hasMoreOutput:%v", c.lastFlush, have, err, c.hasMoreOutput)
	return have, err
}

func (c *decompressor) processReturnValue(ret capi.ZConstant) error {
	// Z_DATA_ERROR indicates that the compressed data was corrupted.
	// Z_NEED_DICT the dictionary needed for decompression is not provided.
	// Z_MEM_ERROR may occur since memory allocation is deferred until inflate() needs it
	// Z_STREAM_ERROR indicates that the stream state was inconsistent
	//                which may happen if the stream was not initialized
	//                Or the appplication is broken and altered the memory of the stream state.
	switch ret {
	case capi.Z_DATA_ERROR, capi.Z_NEED_DICT, capi.Z_MEM_ERROR, capi.Z_STREAM_ERROR:
		reason := fmt.Errorf("zlib: inflate failed with err: %w", capi.ZError(ret))
		return c.endStream(reason)
	}

	if !c.hasMoreOutput {
		// If there is no more output, then it is one of the following cases:
		// - The input is not fully consumed:
		//   - The end of the stream block was encountered before the input is fully consumed. This terminates the decompression. ret should be Z_STREAM_END.
		//   - There should be no other cases where the input is not fully consumed and ret is not Z_STREAM_END.
		// - The input is fully consumed:
		//   - The input ended with the end of the stream block. In that case ret should be Z_STREAM_END.
		//   - The input ended at the end of a block. In that case, the caller should call Feed again with more input.

		// If the input is not fully consumed
		if c.zstream.AvailIn() > 0 {
			// This should not be an actual error.
			if ret == capi.Z_STREAM_END {
				reason := fmt.Errorf("decompression ended but the input was not fully consumed. %w", capi.ZError(ret))
				return c.endStream(reason)
			}

			// This should never happen, but who knows!
			reason := fmt.Errorf("the input was not fully consumed and decompression has not ended (ret != Z_STREAM_END). %w", capi.ZError(ret))
			return c.endStream(reason)
		}

		// The input is fully consumed.
		if ret == capi.Z_STREAM_END {
			err := c.endStream(nil)
			if err != nil {
				return err
			}
			return io.EOF
		}

		if c.lastFlush == Finish {
			// If flush is Z_FINISH, then decompression should have ended with ret = Z_STREAM_END.
			// This indicates that the compressed data is corrupted.
			reason := fmt.Errorf("the end of input was reached (flush=finish) but decompression was not done. The compressed data is probably corrupted. %w", capi.ZError(ret))
			return c.endStream(reason)
		}

		// The input is fully consumed but the stream hasn't ended yet.
		// ret could be Z_BUF_ERROR indicating that more input is needed.
		// or it may not be Z_BUF_ERROR (probably Z_OK) if the current input ended at the end of a block.
		// The full stream is not ended yet and Feed needs to be called with more input.
		// TIE_1: tie this case to a similar one in Consume
		return nil
	}

	// There is still more output to be consumed. The caller should call Consume again.
	// This should coincide with ret = Z_BUF_ERROR which is not a real error.
	// We'll double check to make sure
	// TODO: We happen to get here with ret = Z_OK. I'm not sure why but I'll allow it for now.
	// TODO: We also happen to get here with Z_STREAM_END which means that the stream ended but
	//       we still have more output to be consumed. I'll allow it for now to see what happens
	//       when Consume is called again
	utils.Debug("There is still more output to be consumed. ret: %v", ret)
	if ret != capi.Z_BUF_ERROR && ret != capi.Z_OK && ret != capi.Z_STREAM_END {
		reason := fmt.Errorf("zlib: more output is available but ret is not Z_BUF_ERROR, Z_OK, nor Z_STREAM_END: %w", capi.ZError(ret))
		return c.endStream(reason)
	}
	return nil
}

// inflateEnd needs to be called when ZStream is no longer needed
// Both when the stream is ended and when there is an error.
func (c *decompressor) endStream(reason error) error {
	endRet := c.zstream.InflateEnd()
	c.streamEndHasBeenCalled = true

	c.streamEndReason = reason
	c.streamEndError = processStreamEndError(reason, endRet)
	return c.streamEndError
}

// TODO
// func (c *ZDecompressor2) Reset() err {

// 	ret := c.zstream.DeflateInit2(opts.Level, opts.ZWindowBits(), opts.MemoryLevel, int(opts.Strategy))
// 	if ret != Z_OK {
// 		return nil, ZError(ret)
// 	}
// }
