package main

import (
	"bytes"
	"io"
	"log"

	"github.com/MeenaAlfons/go-zlib/zlib"
	"github.com/MeenaAlfons/go-zlib/zlib/common"
)

func syncCompressWriterFlush(data []byte) []byte {
	var buf bytes.Buffer

	opts := common.DefaultCompressOptions()
	compressWriter, err := zlib.NewCompressWriter(&buf, opts)
	if err != nil {
		log.Fatalf("Error creating compress writer: %v", err)
	}

	if _, err := compressWriter.Write(data[:len(data)/2]); err != nil && err != io.EOF {
		log.Fatalf("Error writing to compress writer: %v", err)
	}

	if err := compressWriter.Flush(); err != nil {
		log.Fatalf("Error flushing to compress writer: %v", err)
	}

	if _, err := compressWriter.Write(data[len(data)/2:]); err != nil && err != io.EOF {
		log.Fatalf("Error writing to compress writer: %v", err)
	}

	if err := compressWriter.Close(); err != nil {
		log.Fatalf("Error closing compress writer: %v", err)
	}

	return buf.Bytes()
}
