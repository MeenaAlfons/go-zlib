package capi

/*
#include <zlib.h>
*/
import "C"

// Maybe It's better to create a new type for return codes, flush values, compression levels, compression strategy, etc.

type ZConstant int

// The following constants are for mappings from zlib.h
// For more details, see http://zlib.net/manual.html#Constants
const (
	// Allowed flush values; see deflate() and inflate() for details
	Z_NO_FLUSH      ZConstant = C.Z_NO_FLUSH
	Z_PARTIAL_FLUSH ZConstant = C.Z_PARTIAL_FLUSH
	Z_SYNC_FLUSH    ZConstant = C.Z_SYNC_FLUSH
	Z_FULL_FLUSH    ZConstant = C.Z_FULL_FLUSH
	Z_FINISH        ZConstant = C.Z_FINISH
	Z_BLOCK         ZConstant = C.Z_BLOCK
	Z_TREES         ZConstant = C.Z_TREES

	// Return codes for the compression/decompression functions.
	Z_OK            ZConstant = C.Z_OK
	Z_STREAM_END    ZConstant = C.Z_STREAM_END
	Z_NEED_DICT     ZConstant = C.Z_NEED_DICT
	Z_ERRNO         ZConstant = C.Z_ERRNO
	Z_STREAM_ERROR  ZConstant = C.Z_STREAM_ERROR
	Z_DATA_ERROR    ZConstant = C.Z_DATA_ERROR
	Z_MEM_ERROR     ZConstant = C.Z_MEM_ERROR
	Z_BUF_ERROR     ZConstant = C.Z_BUF_ERROR
	Z_VERSION_ERROR ZConstant = C.Z_VERSION_ERROR

	// Compression levels
	Z_NO_COMPRESSION      ZConstant = C.Z_NO_COMPRESSION
	Z_BEST_SPEED          ZConstant = C.Z_BEST_SPEED
	Z_BEST_COMPRESSION    ZConstant = C.Z_BEST_COMPRESSION
	Z_DEFAULT_COMPRESSION ZConstant = C.Z_DEFAULT_COMPRESSION

	// Compression strategy; See deflateInit2() for details
	Z_FILTERED         ZConstant = C.Z_FILTERED
	Z_HUFFMAN_ONLY     ZConstant = C.Z_HUFFMAN_ONLY
	Z_RLE              ZConstant = C.Z_RLE
	Z_FIXED            ZConstant = C.Z_FIXED
	Z_DEFAULT_STRATEGY ZConstant = C.Z_DEFAULT_STRATEGY

	// The deflate compression method (the only one supported in this version)
	Z_DEFLATED ZConstant = C.Z_DEFLATED
)
