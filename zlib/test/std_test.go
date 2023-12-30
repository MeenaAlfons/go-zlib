package test

import (
	"bytes"
	"compress/flate"
	"fmt"
	"io"
	"testing"

	"github.com/MeenaAlfons/go-zlib/zlib"
	"github.com/MeenaAlfons/go-zlib/zlib/common"
)

// The whole purpose of this test is to make sure this library is compatible
// with other libraries providing DEFLATE and ZLIB compression and decompression.
func TestAgainstFlate(t *testing.T) {
	tests := []struct {
		name                  string
		operation             func([]byte, common.CompressOptions, func(common.CompressOptions) common.DecompressOptions, testLogger) error
		combinationPredicate  CombinationPredicate
		decompressOptsFactory func(common.CompressOptions) common.DecompressOptions
		expectedError         error
	}{
		{
			name:                  "CompressReaderFlateReader",
			operation:             CompressReaderFlateReader,
			decompressOptsFactory: matchCompressOptions,
			combinationPredicate: And(
				HeaderTypeRawPredicate,
				NilDictionaryPredicate,
			),
		},
		{
			name:                  "FlateWriterDecompressWriter",
			operation:             FlateWriterDecompressWriter,
			decompressOptsFactory: matchCompressOptions,
			combinationPredicate: And(
				HeaderTypeRawPredicate,
				NilDictionaryPredicate,
				FlateInternalWindowBits15Predicate,
			),
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
				test.combinationPredicate,
				test.decompressOptsFactory,
				test.operation,
				test.expectedError,
			)
		})
	}
}

func FlateInternalWindowBits15Predicate(opts common.CompressOptions, sample []byte, _ []byte) bool {
	// Flate has a hard coded window bits of 15
	// This means that it will always consider a window of 1<<15 bytes
	// This means that data compressed by flate cannot be decompressed
	// by another decompresseor that uses window bits of less than 15.
	// unless the data is less than 1<<windowBits bytes so that the
	// flate compressor will not refer to the data outside the window.
	// However, there is a chance that the flate compressor will be
	// compress data longer than 1<<windowBits bytes without referring
	// to the data outside the window, if there is no match found outside
	// the window.
	return opts.WindowBits() == 15 ||
		len(sample) <= 1<<opts.WindowBits()
	// OR there is no repeated data outside the window

	// We'll omit the last condition as it is difficult to check
	// This check is guaranteed to give true for the cases that should pass the test.
	// The compliment of this check is not guaranteed to fail all the tests because in some cases, there will not be any repeated data outside the window.
	// A better data set can be generated to test the difference between those cases expecting a failure and those cases expecting a success.
}

func CompressReaderFlateReader(sampleData []byte, opts common.CompressOptions, decompressOptsFactory func(common.CompressOptions) common.DecompressOptions, log testLogger) error {
	var buf bytes.Buffer
	buf.Write(sampleData)

	compressReader, err := zlib.NewCompressReader(&buf, opts)
	if err != nil {
		return err
	}

	decompressReader := flate.NewReader(compressReader)

	decompressedResult, err := io.ReadAll(decompressReader)
	if err != nil {
		return err
	}

	if !bytes.Equal(decompressedResult, sampleData) {
		return fmt.Errorf("decompressed data is not equal to the original data")
	}

	return nil
}

func FlateWriterDecompressWriter(sampleData []byte, opts common.CompressOptions, decompressOptsFactory func(common.CompressOptions) common.DecompressOptions, log testLogger) error {
	decompressOpts := decompressOptsFactory(opts)
	var buf bytes.Buffer
	decompressWriter, err := zlib.NewDecompressWriter(&buf, decompressOpts)
	if err != nil {
		return err
	}

	flateWriter, _ := flate.NewWriter(decompressWriter, flate.DefaultCompression)
	_, err = flateWriter.Write(sampleData)
	if err != nil {
		return err
	}

	err = flateWriter.Close()
	if err != nil && err != io.EOF {
		return err
	}

	err = decompressWriter.Close()
	if err != nil {
		return err
	}

	decompressedResult := buf.Bytes()

	if !bytes.Equal(decompressedResult, sampleData) {
		return fmt.Errorf("decompressed data is not equal to the original data")
	}

	return nil
}
