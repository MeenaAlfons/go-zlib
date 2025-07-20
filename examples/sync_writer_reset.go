package main

import (
	"bytes"
	"io"
	"log"

	"github.com/MeenaAlfons/go-zlib/zlib"
	"github.com/MeenaAlfons/go-zlib/zlib/common"
)

func syncCompressDecompressWriterReset(data []byte) []byte {
	var buf bytes.Buffer

	compressOpts := common.DefaultCompressOptions()
	compressWriter, err := zlib.NewCompressWriter(&buf, compressOpts)
	if err != nil {
		log.Fatalf("Error creating compress writer: %v", err)
	}

	if _, err := compressWriter.Write(data); err != nil && err != io.EOF {
		log.Fatalf("Error writing to compress writer: %v", err)
	}

	if err := compressWriter.Close(); err != nil {
		log.Fatalf("Error closing compress writer: %v", err)
	}
	compressed := buf.Bytes()
	buf.Reset() // Reset the buffer for decompression

	decompressOpts := common.DefaultDecompressOptions()
	decompressWriter, err := zlib.NewDecompressWriter(&buf, decompressOpts)
	if err != nil {
		log.Fatalf("Error creating decompress writer: %v", err)
	}

	if _, err := decompressWriter.Write(compressed); err != nil && err != io.EOF {
		log.Fatalf("Error writing to decompress writer: %v", err)
	}

	if err := decompressWriter.Close(); err != nil {
		log.Fatalf("Error closing decompress writer: %v", err)
	}
	decompressed := buf.Bytes()
	buf.Reset() // Reset the buffer for the next operation

	for i := 0; i < 1000; i++ {
		compressWriter.Reset(&buf)
		if _, err := compressWriter.Write(decompressed); err != nil && err != io.EOF {
			log.Fatalf("Error writing to compress writer: %v", err)
		}

		if err := compressWriter.Close(); err != nil {
			log.Fatalf("Error closing compress writer: %v", err)
		}
		compressed = buf.Bytes()
		buf.Reset() // Reset the buffer for decompression

		decompressWriter.Reset(&buf)
		if _, err := decompressWriter.Write(compressed); err != nil && err != io.EOF {
			log.Fatalf("Error writing to decompress writer: %v", err)
		}

		if err := decompressWriter.Close(); err != nil {
			log.Fatalf("Error closing decompress writer: %v", err)
		}
		decompressed = buf.Bytes()
		buf.Reset() // Reset the buffer for the next operation
	}

	return decompressed
}
