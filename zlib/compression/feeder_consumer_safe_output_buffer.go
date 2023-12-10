package compression

// When the stream arrives at the end of a buffer, its internal state would refer to a position past the end of the buffer.
// This results in error: "found pointer to free object".
// To avoid this case, we reserve one byte at the end of the buffer so that the final state will not point past the end of the buffer.
//
// feederConsumerSafeOutputBuffer is a wrapper around FeederConsumer that accepts a buffer and reserves the last bytes of the output buffer for memory safety reasons.
// The wrapped FeederConsumer will be called with an output buffer that is one byte smaller than the allocated output buffer.
//
// If feederConsumerSafeOutputBuffer is called with a buffer with size 1, it will call the wrapped FeederConsumer with a new buffer with capacity 2 and size 1 and copy the result to the original buffer after the operation.
//
// Note that the input buffer is not yet protected by this wrapper.
type feederConsumerSafeOutputBuffer struct {
	feederConsumer FeederConsumer
}

func newFeederConsumerSafeOutputBuffer(feederConsumer FeederConsumer) FeederConsumer {
	return &feederConsumerSafeOutputBuffer{
		feederConsumer: feederConsumer,
	}
}

// TODO: input needs to be guarded as well!
func (c *feederConsumerSafeOutputBuffer) Feed(input []byte, flush Flush, outputBuffer []byte) (int, error) {
	if len(outputBuffer) > 1 {
		return c.feederConsumer.Feed(input, flush, outputBuffer[:len(outputBuffer)-1])
	}

	newOutputBuffer := make([]byte, 2)
	have, err := c.feederConsumer.Feed(input, flush, newOutputBuffer[:len(newOutputBuffer)-1])
	copy(outputBuffer, newOutputBuffer)
	return have, err
}

func (c *feederConsumerSafeOutputBuffer) Consume(outputBuffer []byte) (int, error) {
	if len(outputBuffer) > 1 {
		return c.feederConsumer.Consume(outputBuffer[:len(outputBuffer)-1])
	}

	newOutputBuffer := make([]byte, 2)
	have, err := c.feederConsumer.Consume(newOutputBuffer[:len(newOutputBuffer)-1])
	copy(outputBuffer, newOutputBuffer)
	return have, err
}

func (c *feederConsumerSafeOutputBuffer) CanCallConsume() bool {
	return c.feederConsumer.CanCallConsume()
}

func (c *feederConsumerSafeOutputBuffer) IsDoneWithReason() (bool, error) {
	return c.feederConsumer.IsDoneWithReason()
}
