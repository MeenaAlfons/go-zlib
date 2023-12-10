# go-zlib

Let's you use `zlib` in Go. Direct fully-featured API bindings to `zlib`.

It improves on the current standard libraries [zlib](https://pkg.go.dev/compress/zlib) and [deflate](https://pkg.go.dev/compress/deflate) by supporting all the possible options for `zlib`.

Easy to use `Writer` and `Reader` interfaces in addition to direct and memory-safe access to zlib low-level methods.

## Why another zlib library in Go?


- **Hidden configurations:** The standard libraries [zlib](https://pkg.go.dev/compress/zlib) and [deflate](https://pkg.go.dev/compress/deflate) hide most of the configurations of the underlying compression algorithm. A library like [4kills/go-zlib](github.com/4kills/go-zlib) exposes some of those configurations like `compression strategies` and `compression levels` but still hides other parameters like `WindowBits` and `MemoryLevel`. This library exposes all these configurations in addition to giving more control on memory consumption via the `BufferSize` parameter.
- **Limited IO interfaces:** The prevously mentioned libraries implement `Writer` interface for compression and `Reader` interface for decompression. However, there are cases where a compression reader or a decompression writer are needed. This library supports both interfaces for both compression and decompression.
- **Access to lower level APIs:** The previously mentioned libraries only support `Reader` and `Writer` interfaces. In addition to supporting these interfaces, this library gives access to `compressor` and `decompressor` via `FeederConsumer` interface which allows incremental feeding of data and consumption of results. In addition to that, it gives direct access to zlib `zstream` and its methods.

## Installation

```sh
go get github.com/MeenaAlfons/go-zlib
```

### Prerequisites

This library assumes that `zlib` is already installed and its include and library files are accessable to build tools.

Many operating systems come with `zlib` already installed. If you need to install `zlib`, here is a list of commands to help:
- macos: `brew install zlib`
- debian: `apt install zlib`
- alpine: `apk add zlib`
- windows: download binary [here](https://gnuwin32.sourceforge.net/packages/zlib.htm)


## Usage

Compression and decompression can be plugged into any pipeline via `Writer` and `Reader` interfaces. Compression and decompression are supported on both writing and reading sides. Here are the possilities:
- `NewCompressReader`
- `NewCompressWriter`
- `NewDecompressReader`
- `NewDecompressWriter`

Many examples can be found in [examples](examples) directory. Here is one example:

```go
data := []byte("Hello World!")
r := bytes.NewReader(data)

opts := common.DefaultCompressOptions()
compressorReader, err := zlib.NewCompressReader(r, opts)
if err != nil {
    // Error creating compressor reader
}

compressedData, err := io.ReadAll(compressorReader)
if err != nil {
    // Error reading from compressor reader
}
return compressedData
```

## Development

Run tests
```sh
go test ./...
```

Run tests with debug logs
```sh
go test ./... -tags debug
```

## License

[MIT License](LICENSE)

## Acknowledgments

* [zlib](https://www.zlib.net/) is written in C by Jean-loup Gailly and Mark Adler.

