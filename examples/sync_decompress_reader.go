package main

import (
	"bytes"
	"io"
	"log"

	"github.com/MeenaAlfons/go-zlib/zlib"
	"github.com/MeenaAlfons/go-zlib/zlib/common"
)

func syncDecompressReader(data []byte) []byte {
	r := bytes.NewReader(data)

	opts := common.DefaultDecompressOptions()
	deompressReader, err := zlib.NewDecompressReader(r, opts)
	if err != nil {
		log.Fatalf("Error creating decompress reader: %v", err)
	}

	decompressedData, err := io.ReadAll(deompressReader)
	if err != nil {
		log.Fatalf("Error reading from decompress reader: %v", err)
	}
	return decompressedData
}
