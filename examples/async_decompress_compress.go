package main

import (
	"io"
	"log"
	"sync"

	"github.com/MeenaAlfons/go-zlib/zlib"
	"github.com/MeenaAlfons/go-zlib/zlib/common"
)

func asynchronousDecompressAndCompress(data []byte) []byte {
	pipeReader, pipeWriter := io.Pipe()

	decompressWriter, err := zlib.NewDecompressWriter(pipeWriter, common.DefaultDecompressOptions())
	if err != nil {
		log.Fatalf("Error creating decompressor writer: %v", err)
	}

	compressReader, err := zlib.NewCompressReader(pipeReader, common.DefaultCompressOptions())
	if err != nil {
		log.Fatalf("Error creating compressor reader: %v", err)
	}

	var compressReaderErr error
	var compressedData []byte
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		compressedData, compressReaderErr = io.ReadAll(compressReader)
	}()

	chunkSize := 2
	for _, chunk := range chunked(data, chunkSize) {
		if _, err := decompressWriter.Write(chunk); err != nil && err != io.EOF {
			log.Fatalf("Error writing to decompress writer: %v", err)
		}
		if err := decompressWriter.Flush(); err != nil {
			log.Fatalf("Error flushing decompress writer: %v", err)
		}
	}

	if err := decompressWriter.Close(); err != nil {
		log.Fatalf("Error closing decompress writer: %v", err)
	}

	if err := pipeWriter.Close(); err != nil {
		log.Fatalf("Error closing pipe writer: %v", err)
	}

	wg.Wait()

	if compressReaderErr != nil && compressReaderErr != io.EOF {
		log.Fatalf("Error reading from compress reader: %v", compressReaderErr)
	}

	return compressedData
}
