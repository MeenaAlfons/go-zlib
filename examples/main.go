package main

import "log"

func main() {
	data := []byte("Hello World! That was such a long nap!")

	compressed := syncCompressReader(data)
	decompressed := syncDecompressReader(compressed)

	compressed2 := syncCompressWriter(data)
	decompressed2 := syncDecompressWriter(compressed2)

	compressed3 := syncCompressWriterFlush(data)
	decompressed3 := syncDecompressWriterFlush(compressed3)

	compressed4 := asynchronousDecompressAndCompress(compressed3)
	decompressed4 := syncDecompressWriter(compressed4)

	decompressed5 := asynchronousCompressAndDecompress(data)

	log.Printf(`
		data: %v

		compressed: %v
		compressed2: %v
		compressed3: %v
		compressed4: %v

		decompressed: %v
		decompressed2: %v
		decompressed3: %v
		decompressed4: %v
		decompressed5: %v
		`,
		data,
		compressed,
		compressed2,
		compressed3,
		compressed4,
		decompressed,
		decompressed2,
		decompressed3,
		decompressed4,
		decompressed5,
	)
}
