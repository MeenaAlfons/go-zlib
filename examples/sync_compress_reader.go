package main

import (
	"bytes"
	"io"
	"log"

	"github.com/MeenaAlfons/go-zlib/zlib"
	"github.com/MeenaAlfons/go-zlib/zlib/common"
)

func syncCompressReader(data []byte) []byte {
	r := bytes.NewReader(data)

	opts := common.DefaultCompressOptions()
	compressReader, err := zlib.NewCompressReader(r, opts)
	if err != nil {
		log.Fatalf("Error creating compress reader: %v", err)
	}

	compressedData, err := io.ReadAll(compressReader)
	if err != nil {
		log.Fatalf("Error reading from compress reader: %v", err)
	}
	return compressedData
}
