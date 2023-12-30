package test

import (
	"bytes"
	"fmt"
	"io"
	"sync"
	"testing"

	"github.com/MeenaAlfons/go-zlib/zlib"
	"github.com/MeenaAlfons/go-zlib/zlib/common"
)

// Whenever we test, we need to test either compression or decompression.
// It's best if we can test one at a time so that we are sure it works against
// previously known outcome.

// However, it's quite difficult to specify compressed outcome.
// So, we will be testing them against each other.
// Given that we are testing a port of another library, we are not interested in testing the actual outcome.
// We are rather interested in testing that the outcome is the same as the original library.
// We'll do that by feeding it again through the transformation.

// It's still not the best because a bug in one direction could have a complementary bug in the other direction.
// But it's better than nothing.

// We will be testing against the following data:
// 1. predefined data for which we can specify the outcome (even the compressed one).
// 2. random sequences of bytes with varying lengths.

// We will be testing 4 components:
// 1. CompressorReader
// 2. CompressorWriter
// 3. DecompressorReader
// 4. DecompressorWriter

// We will be testing many combinations of options. The options are:
// 1. Level
// 2. WindowBits
// 3. Header
// 4. MemoryLevel
// 5. Strategy
// 6. BufferSize

// TODO Create a test to verify the functionality of Flush vs Close on Writers.
// TODO create a test to veify that windowbit can be set to 0 to request that  the window size in the zlib header of the compressed stream will be used.
// TODO Add test cases for the error or failure cases.

const REPEAT_COUNT = 1

func TestZlib(t *testing.T) {
	tests := []struct {
		name                  string
		operation             func([]byte, common.CompressOptions, func(common.CompressOptions) common.DecompressOptions, testLogger) error
		decompressOptsFactory func(opts common.CompressOptions) common.DecompressOptions
		expectedError         error
	}{
		{
			name:                  "DecompressorWriterCompressReader",
			operation:             DecompressorWriterCompressReader,
			decompressOptsFactory: matchCompressOptions,
			expectedError:         nil,
		},
		{
			name:                  "CompressorWriterDecompressReader",
			operation:             CompressorWriterDecompressReader,
			decompressOptsFactory: matchCompressOptions,
			expectedError:         nil,
		},
	}

	samples := getDataSamples()
	combinations := getCompressOptionsCombinations()
	dictionaryFactories := getDictionaryFactories()

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			RunTestOnCombinations(
				t,
				samples,
				combinations,
				dictionaryFactories,
				REPEAT_COUNT,
				AllCombinationsPredicate,
				test.decompressOptsFactory,
				test.operation,
				test.expectedError,
			)
		})
	}
}

func WithDictionary(operation func(*testing.T, sample, common.CompressOptions, func(common.CompressOptions) common.DecompressOptions) error) func(*testing.T, sample, common.CompressOptions, func(common.CompressOptions) common.DecompressOptions) error {
	return func(t *testing.T, sample sample, compressOpts common.CompressOptions, decompressOptsFactory func(common.CompressOptions) common.DecompressOptions) error {
		dictionary := []byte("This is a dictionary")
		newCompressionOpts := common.DefaultCompressOptions().WithBufferSize(compressOpts.BufferSize()).WithLevel(compressOpts.Level()).WithWindowBits(compressOpts.WindowBits()).WithHeader(compressOpts.Header()).WithMemoryLevel(compressOpts.MemoryLevel()).WithStrategy(compressOpts.Strategy()).WithInitialDictionary(dictionary)
		return operation(t, sample, newCompressionOpts, decompressOptsFactory)
	}
}

func DecompressorWriterCompressReader(sampleData []byte, opts common.CompressOptions, decompressOptsFactory func(common.CompressOptions) common.DecompressOptions, log testLogger) error {
	decompressOpts := decompressOptsFactory(opts)

	compressed, err := synchronousCompressReader(log, sampleData, opts)
	if err != nil {
		return err
	}
	log.Logf("Compressed data 1: %x", compressed)

	compressed2, decompressed, err := asynchronousDecompressAndCompress(log, compressed, opts, decompressOpts)
	if err != nil {
		return err
	}
	log.Logf("Decompressed data 1: %x", decompressed)
	log.Logf("Compressed data 2: %x", compressed2)

	// compressed and compressed2 could be different.
	// So, we can't compare them.
	// Instead, we'll decompress compressed2 to make sure it is correct.
	decompressed2, err := synchronousDecompressWriter(log, compressed2, decompressOpts)
	if err != nil {
		return err
	}
	log.Logf("Decompressed data 2: %x", compressed2)

	if !bytes.Equal(decompressed, sampleData) {
		return fmt.Errorf("decompressed != original data\n decompressed: %x\n original: %x", decompressed, sampleData)
	}

	if !bytes.Equal(decompressed2, sampleData) {
		return fmt.Errorf("decompressed2 != original data\n decompressed: %x\n original: %x", decompressed, sampleData)
	}

	return nil
}

func CompressorWriterDecompressReader(sampleData []byte, opts common.CompressOptions, decompressOptsFactory func(common.CompressOptions) common.DecompressOptions, log testLogger) error {
	decompressOpts := decompressOptsFactory(opts)

	compressed, decompressed, err := asynchronousCompressAndDecompress(log, sampleData, opts, decompressOpts)
	if err != nil {
		return err
	}
	log.Logf("Decompressed data 1: %x", decompressed)
	log.Logf("Compressed data 1: %x", compressed)

	compressed2, err := synchronousCompressWriter(log, decompressed, opts)
	if err != nil {
		return err
	}
	log.Logf("Compressed data 2: %x", compressed2)

	// compressed and compressed2 are expected to be different because we use SynchFlush followed by finish in the asynchronous version
	// while we use finish only in the synchronous version.
	// So, we can't compare them.
	// Instead, we'll decompress compressed2 to make sure it is correct.
	decompressed2, err := synchronousDecompressWriter(log, compressed2, decompressOpts)
	if err != nil {
		return err
	}
	log.Logf("Decompressed data 2: %x", compressed2)

	if !bytes.Equal(decompressed, sampleData) {
		return fmt.Errorf("decompressed != original data\n decompressed: %x\n original: %x", decompressed, sampleData)
	}

	if !bytes.Equal(decompressed2, sampleData) {
		return fmt.Errorf("decompressed2 != original data\n decompressed: %x\n original: %x", decompressed, sampleData)
	}

	return nil
}

// Test CompressorReader synchronously when the full input is already available.
func synchronousCompressReader(log testLogger, decompressed []byte, opts common.CompressOptions) ([]byte, error) {
	var buf bytes.Buffer
	if _, err := buf.Write(decompressed); err != nil {
		return nil, fmt.Errorf("Error writing to buffer: %w", err)
	}

	compressorReader, err := zlib.NewCompressReader(&buf, opts)
	if err != nil {
		return nil, fmt.Errorf("Error creating compressor reader: %w", err)
	}

	outCompressedData, err := io.ReadAll(compressorReader)
	if err != nil {
		return nil, fmt.Errorf("Error reading from compressor reader: %w", err)
	}
	return outCompressedData, nil
}

// Test CompressorWriter synchronously when the full input is already available.
func synchronousCompressWriter(log testLogger, decompressed []byte, opts common.CompressOptions) ([]byte, error) {
	var buf bytes.Buffer
	compressorWriter, err := zlib.NewCompressWriter(&buf, opts)
	if err != nil {
		return nil, fmt.Errorf("Error creating compressor writer: %w", err)
	}

	log.Log("Before writing to compressor writer")
	if _, err := compressorWriter.Write(decompressed); err != nil && err != io.EOF {
		return nil, fmt.Errorf("Error writing to compressor writer: %w\n compressed so far: %x", err, buf.Bytes())
	}
	log.Log("After writing to compressor writer")

	if err := compressorWriter.Flush(); err != nil {
		return nil, fmt.Errorf("Error flushing compressor writer: %w", err)
	}
	log.Log("After flushing to compressor writer")

	if err := compressorWriter.Close(); err != nil {
		return nil, fmt.Errorf("Error closing compressor writer: %w", err)
	}
	log.Log("After closing to compressor writer")

	return buf.Bytes(), nil
}

func synchronousDecompressWriter(log testLogger, compressed []byte, opts common.DecompressOptions) ([]byte, error) {
	var buf bytes.Buffer
	decompressorWriter, err := zlib.NewDecompressWriter(&buf, opts)
	if err != nil {
		return nil, fmt.Errorf("Error creating decompressor writer: %w", err)
	}

	log.Log("Before writing to decompressor writer")
	if _, err := decompressorWriter.Write(compressed); err != nil && err != io.EOF {
		return nil, fmt.Errorf("Error writing to decompressor writer: %w\n decompressed so far: %x", err, buf.Bytes())
	}
	log.Log("After writing to decompressor writer")

	if err := decompressorWriter.Flush(); err != nil {
		return nil, fmt.Errorf("Error flushing decompressor writer: %w", err)
	}
	log.Log("After flushing to decompressor writer")

	if err := decompressorWriter.Close(); err != nil {
		return nil, fmt.Errorf("Error closing decompressor writer: %w", err)
	}
	log.Log("After closing to decompressor writer")

	return buf.Bytes(), nil
}

// Test the asynchronous behavior of DecompressorWriter and CompressorReader
// by connecting them through a pipe.
// A side buffer is used to store the decompressed data.
func asynchronousDecompressAndCompress(log testLogger, compressed []byte, opts common.CompressOptions, decompressOpts common.DecompressOptions) (compressedResult []byte, decompressedResult []byte, err error) {
	var decompressedBuffer bytes.Buffer
	defer func() {
		decompressedResult = decompressedBuffer.Bytes()
	}()

	pipeReader, pipeWriter := io.Pipe()
	multiWriter := io.MultiWriter(pipeWriter, &decompressedBuffer)

	decompressorWriter, err := zlib.NewDecompressWriter(multiWriter, decompressOpts)
	if err != nil {
		return nil, nil, fmt.Errorf("Error creating decompressor writer: %w", err)
	}

	compressorReader, err := zlib.NewCompressReader(pipeReader, opts)
	if err != nil {
		return nil, nil, fmt.Errorf("Error creating compressor reader: %w", err)
	}

	var compressReaderErr error
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		compressedResult, compressReaderErr = io.ReadAll(compressorReader)
	}()

	log.Log("Before writing to decompressor writer")
	if _, err := decompressorWriter.Write(compressed); err != nil && err != io.EOF {
		return nil, nil, fmt.Errorf("Error writing to decompressor writer: %w\n decompressed so far: %x", err, decompressedBuffer.Bytes())
	}
	log.Log("After writing to decompressor writer")

	if err := decompressorWriter.Flush(); err != nil {
		return nil, nil, fmt.Errorf("Error flushing decompressor writer: %w", err)
	}
	log.Log("After flushing to decompressor writer")

	if err := decompressorWriter.Close(); err != nil {
		return nil, nil, fmt.Errorf("Error closing decompressor writer: %w", err)
	}
	log.Log("After closing to decompressor writer")

	if err := pipeWriter.Close(); err != nil {
		return nil, nil, fmt.Errorf("Error closing pipe writer: %w", err)
	}

	wg.Wait()

	if compressReaderErr != nil && compressReaderErr != io.EOF {
		return nil, nil, fmt.Errorf("Error reading from compressor reader: %w", compressReaderErr)
	}

	return compressedResult, decompressedResult, nil
}

// Test the asynchronous behavior of CompressorWriter and DecompressorReader
// by connecting them through a pipe.
// A side buffer is used to store the decompressed data.
func asynchronousCompressAndDecompress(log testLogger, decompressed []byte, opts common.CompressOptions, decompressOpts common.DecompressOptions) (compressedResult []byte, decompressedResult []byte, err error) {
	var compressedBuffer bytes.Buffer
	defer func() {
		compressedResult = compressedBuffer.Bytes()
	}()

	pipeReader, pipeWriter := io.Pipe()
	multiWriter := io.MultiWriter(pipeWriter, &compressedBuffer)

	compressorWriter, err := zlib.NewCompressWriter(multiWriter, opts)
	if err != nil {
		return nil, nil, fmt.Errorf("Error creating compressor writer: %w", err)
	}

	decompressorReader, err := zlib.NewDecompressReader(pipeReader, decompressOpts)
	if err != nil {
		return nil, nil, fmt.Errorf("Error creating decompressor reader: %w", err)
	}

	var decompressReaderErr error
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		decompressedResult, decompressReaderErr = io.ReadAll(decompressorReader)
	}()

	log.Log("Before writing to compressor writer")
	if _, err := compressorWriter.Write(decompressed); err != nil && err != io.EOF {
		return nil, nil, fmt.Errorf("Error writing to compressor writer: %w\n compressed so far: %x", err, compressedBuffer.Bytes())
	}
	log.Log("After writing to compressor writer")

	if err := compressorWriter.Flush(); err != nil {
		return nil, nil, fmt.Errorf("Error flushing compressor writer: %w", err)
	}
	log.Log("After flushing compressor writer")

	if err := compressorWriter.Close(); err != nil {
		return nil, nil, fmt.Errorf("Error closing compressor writer: %w", err)
	}
	log.Log("After closing compressor writer")

	if err := pipeWriter.Close(); err != nil {
		return nil, nil, fmt.Errorf("Error closing pipe writer: %w", err)
	}

	wg.Wait()

	if decompressReaderErr != nil && decompressReaderErr != io.EOF {
		return nil, nil, fmt.Errorf("Error reading from decompressor reader: %w", decompressReaderErr)
	}

	return compressedResult, decompressedResult, nil
}

func repeat(n int, f func(index int)) {
	for i := 0; i < n; i++ {
		f(i)
	}
	// runtime.GC()
}

type testLogger interface {
	Log(args ...interface{})
	Logf(format string, args ...interface{})
}

func copyCompressOptions(opts common.CompressOptions) common.CompressOptions {
	return common.DefaultCompressOptions().WithBufferSize(opts.BufferSize()).WithLevel(opts.Level()).WithWindowBits(opts.WindowBits()).WithHeader(opts.Header()).WithMemoryLevel(opts.MemoryLevel()).WithStrategy(opts.Strategy()).WithInitialDictionary(opts.InitialDictionary())
}

func matchCompressOptions(opts common.CompressOptions) common.DecompressOptions {
	return common.DefaultDecompressOptions().WithWindowBits(opts.WindowBits()).WithHeader(opts.Header()).WithInitialDictionary(opts.InitialDictionary())
}
