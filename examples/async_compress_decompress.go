package main

import (
	"io"
	"log"
	"sync"

	"github.com/MeenaAlfons/go-zlib/zlib"
	"github.com/MeenaAlfons/go-zlib/zlib/common"
)

func asynchronousCompressAndDecompress(data []byte) []byte {
	pipeReader, pipeWriter := io.Pipe()

	compressWriter, err := zlib.NewCompressWriter(pipeWriter, common.DefaultCompressOptions())
	if err != nil {
		log.Fatalf("Error creating compressor writer: %v", err)
	}

	decompressReader, err := zlib.NewDecompressReader(pipeReader, common.DefaultDecompressOptions())
	if err != nil {
		log.Fatalf("Error creating decompressor reader: %v", err)
	}

	var decompressReaderErr error
	var decompressedData []byte
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		decompressedData, decompressReaderErr = io.ReadAll(decompressReader)
	}()

	chunkSize := 2
	for _, chunk := range chunked(data, chunkSize) {
		if _, err := compressWriter.Write(chunk); err != nil && err != io.EOF {
			log.Fatalf("Error writing to compress writer: %v", err)
		}
		if err := compressWriter.Flush(); err != nil {
			log.Fatalf("Error flushing compress writer: %v", err)
		}
	}

	if err := compressWriter.Close(); err != nil {
		log.Fatalf("Error closing compress writer: %v", err)
	}

	if err := pipeWriter.Close(); err != nil {
		log.Fatalf("Error closing pipe writer: %v", err)
	}

	wg.Wait()

	if decompressReaderErr != nil && decompressReaderErr != io.EOF {
		log.Fatalf("Error reading from decompress reader: %v", decompressReaderErr)
	}

	return decompressedData
}
