package compression

// flush has two meanings:
// - It can be used to force flushing as much output as possible, like concluding the compression of the current input allowing this block to be decompressed independently from the next block.
// - It can be used to indicate that the stream has ended and no more input will be fed.
type Flush int

const (
	NoFlush   Flush = 0
	SyncFlush Flush = 2
	Finish    Flush = 3
)

// FeederConsumer is an interface that allows feeding input and consuming output.
// It represents a component that takes input and produces output in a streaming fashion.
// It is used to implement both compression and decompression.
type FeederConsumer interface {
	// Feed feeds input to the stream. It returns the number of bytes written to the output buffer.
	// If the output buffer is not large enough, it writes as much as possible to the output buffer.
	// The rest of the output needs to be consumed by caling Consume.
	// flush can be used to force flushing as much output as possible, or to indicate the end of the stream.
	Feed(input []byte, flush Flush, outputBuffer []byte) (int, error)

	// Consume consumes output from the stream. It returns the number of bytes written to the output buffer.
	Consume(outputBuffer []byte) (int, error)

	// CanCallConsume returns true if there is output that needs to be consumed by calling Consume.
	CanCallConsume() bool

	// IsDoneWithReason returns true if the stream has ended.
	// If the stream has ended because of an error, it returns the error.
	IsDoneWithReason() (bool, error)
}
