package main

import (
	"bytes"
	"io"
	"log"

	"github.com/MeenaAlfons/go-zlib/zlib"
	"github.com/MeenaAlfons/go-zlib/zlib/common"
)

func syncDecompressWriterFlush(data []byte) []byte {
	var buf bytes.Buffer

	opts := common.DefaultDecompressOptions()
	decompressWriter, err := zlib.NewDecompressWriter(&buf, opts)
	if err != nil {
		log.Fatalf("Error creating decompress writer: %v", err)
	}

	if _, err := decompressWriter.Write(data[:len(data)/2]); err != nil && err != io.EOF {
		log.Fatalf("Error writing to decompress writer: %v", err)
	}

	if err := decompressWriter.Flush(); err != nil {
		log.Fatalf("Error flushing to decompress writer: %v", err)
	}

	if _, err := decompressWriter.Write(data[len(data)/2:]); err != nil && err != io.EOF {
		log.Fatalf("Error writing to decompress writer: %v", err)
	}

	if err := decompressWriter.Close(); err != nil {
		log.Fatalf("Error closing decompress writer: %v", err)
	}

	return buf.Bytes()
}
