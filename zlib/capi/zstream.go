package capi

/*
#include <zlib.h>

// The following functions are dummy mappings from zlib.h
// The functions defined here can be called from Go code.

int DeflateInit(z_streamp strm, int level) {
	return deflateInit(strm, level);
}

int DeflateInit2(z_streamp strm, int level,
                                 int method,
                                 int windowBits,
                                 int memLevel,
                                 int strategy) {
	return deflateInit2(strm, level, method, windowBits, memLevel, strategy);
}

int InflateInit(z_streamp strm) {
	return inflateInit(strm);
}

int InflateInit2(z_streamp strm, int windowBits) {
	return inflateInit2(strm, windowBits);
}

int DeflateSetDictionary(z_streamp strm, const Bytef *dictionary,
						 uInt  dictLength) {
	return deflateSetDictionary(strm, dictionary, dictLength);
}

int InflateSetDictionary(z_streamp strm, const Bytef *dictionary,
						 uInt  dictLength) {
	return inflateSetDictionary(strm, dictionary, dictLength);
}

int DeflateBound(z_streamp strm, int sourceLen) {
	return deflateBound(strm, sourceLen);
}

int Inflate(z_streamp strm, int flush) {
	return inflate(strm, flush);
}

int Deflate(z_streamp strm, int flush) {
	return deflate(strm, flush);
}

int DeflateEnd(z_streamp strm) {
	return deflateEnd(strm);
}

int InflateEnd(z_streamp strm) {
	return inflateEnd(strm);
}
*/
import "C"

import (
	"runtime"

	"github.com/MeenaAlfons/go-zlib/zlib/utils"
)

type ZStream interface {
	InflateInit() ZConstant
	InflateInit2(windowBits int) ZConstant
	DeflateInit(level int) ZConstant
	DeflateInit2(level, windowBits, memLevel, strategy int) ZConstant

	DeflateSetDictionary(dictionary []byte) ZConstant
	InflateSetDictionary(dictionary []byte) ZConstant

	DeflateEnd() ZConstant
	InflateEnd() ZConstant

	SetInput(in []byte)
	SetOutput(out []byte)

	Deflate(flush ZConstant) ZConstant
	Inflate(flush ZConstant) ZConstant

	ProducedOutput() int
	OutputBufferIsFull() bool
	AvailIn() int

	DeflateBound(sourceLength int) int
}

// NewZStream creates a new ZStream representing a C z_stream
func NewZStream() ZStream {
	z := &zstream{}
	utils.Debug("NewZStream %p", z)

	return z
}

// Make sure that zstream implements ZStream
var _ ZStream = (*zstream)(nil)

type zstream struct {
	strm C.z_stream

	in  []byte
	out []byte
}

// InflateInit initializes the internal stream state for decompression.
// For more details, see http://zlib.net/manual.html#Basic
func (z *zstream) InflateInit() ZConstant {
	z.strm.zalloc = nil
	z.strm.zfree = nil
	z.strm.opaque = nil
	z.strm.next_in = nil
	z.strm.avail_in = 0

	pinner := runtime.Pinner{}
	pinner.Pin(&z.strm)
	defer pinner.Unpin()

	return ZConstant(C.InflateInit(&z.strm))
}

// InflateInit2 initializes the internal stream state for decompression.
// For more details, see http://zlib.net/manual.html#Advanced
func (z *zstream) InflateInit2(windowBits int) ZConstant {
	z.strm.zalloc = nil
	z.strm.zfree = nil
	z.strm.opaque = nil
	z.strm.next_in = nil
	z.strm.avail_in = 0

	pinner := runtime.Pinner{}
	pinner.Pin(&z.strm)
	defer pinner.Unpin()

	return ZConstant(C.InflateInit2(&z.strm, C.int(windowBits)))
}

// DeflateInit initializes the internal stream state for compression.
// For more details, see http://zlib.net/manual.html#Basic
func (z *zstream) DeflateInit(level int) ZConstant {
	z.strm.zalloc = nil
	z.strm.zfree = nil
	z.strm.opaque = nil

	pinner := runtime.Pinner{}
	pinner.Pin(&z.strm)
	defer pinner.Unpin()

	return ZConstant(C.DeflateInit(&z.strm, C.int(level)))
}

// DeflateInit2 initializes the internal stream state for compression.
// For more details, see http://zlib.net/manual.html#Advanced
func (z *zstream) DeflateInit2(level, windowBits, memLevel, strategy int) ZConstant {
	z.strm.zalloc = nil
	z.strm.zfree = nil
	z.strm.opaque = nil

	pinner := runtime.Pinner{}
	pinner.Pin(&z.strm)
	defer pinner.Unpin()

	return ZConstant(C.DeflateInit2(&z.strm, C.int(level), C.int(Z_DEFLATED), C.int(windowBits), C.int(memLevel), C.int(strategy)))
}

// DeflateSetDictionary initializes the compression dictionary.
// For more details, see http://zlib.net/manual.html#Advanced
func (z *zstream) DeflateSetDictionary(dictionary []byte) ZConstant {
	// Copy the dictionary for two reasons:
	// - Not depending on the caller to keep the dictionary alive
	// - Providing one more byte at the end of the dictionary to avoid
	//   the case where the stream ends at the end of the dictionary which
	//   would result in an internal state that points past the end of the
	//   dictionary and causes an error "found pointer to free object".
	dict := make([]byte, len(dictionary)+1)
	copy(dict, dictionary)
	utils.Debug("DeflateSetDictionary %p dict: %p len(dict): %d, dictionary: %p, len(dictionary): %d (*C.Bytef)(&dict[0]): %p", z, dict, len(dict), dictionary, len(dictionary), (*C.Bytef)(&dict[0]))

	pinner := runtime.Pinner{}
	pinner.Pin(&z.strm)
	defer pinner.Unpin()

	return ZConstant(C.DeflateSetDictionary(&z.strm, (*C.Bytef)(&dict[0]), C.uInt(len(dictionary))))
}

// InflateSetDictionary initializes the decompression dictionary.
// For more details, see http://zlib.net/manual.html#Advanced
func (z *zstream) InflateSetDictionary(dictionary []byte) ZConstant {
	// Copy the dictionary for two reasons:
	// - Not depending on the caller to keep the dictionary alive
	// - Providing one more byte at the end of the dictionary to avoid
	//   the case where the stream ends at the end of the dictionary which
	//   would result in an internal state that points past the end of the
	//   dictionary and causes an error "found pointer to free object".
	dict := make([]byte, len(dictionary)+1)
	copy(dict, dictionary)
	utils.Debug("InflateSetDictionary %p dict: %p len(dict): %d, dictionary: %p, len(dictionary): %d (*C.Bytef)(&dict[0]): %p", z, dict, len(dict), dictionary, len(dictionary), (*C.Bytef)(&dict[0]))

	pinner := runtime.Pinner{}
	pinner.Pin(&z.strm)
	defer pinner.Unpin()

	return ZConstant(C.InflateSetDictionary(&z.strm, (*C.Bytef)(&dict[0]), C.uInt(len(dictionary))))
}

// DeflateBound returns an upper bound on the compressed size after deflation of sourceLen bytes.
// For more details, see http://zlib.net/manual.html#Advanced
func (z *zstream) DeflateBound(sourceLength int) int {
	pinner := runtime.Pinner{}
	pinner.Pin(&z.strm)
	defer pinner.Unpin()

	return int(C.DeflateBound(&z.strm, C.int(sourceLength)))
}

// DeflateEnd frees all dynamically allocated data structures for this stream.
// For more details, see http://zlib.net/manual.html#Basic
func (z *zstream) DeflateEnd() ZConstant {
	z.SetInput(nil)
	z.SetOutput(nil)

	pinner := z.pin()
	defer func() {
		pinner.Unpin()
		utils.Debug("DeflateEnd Unpinned")
	}()
	utils.Debug("DeflateEnd %p next_int: %p out: next_out %p", z, z.strm.next_in, z.strm.next_out)

	return ZConstant(C.DeflateEnd(&z.strm))
}

// InflateEnd frees all dynamically allocated data structures for this stream.
// For more details, see http://zlib.net/manual.html#Basic
func (z *zstream) InflateEnd() ZConstant {

	z.SetInput(nil)
	z.SetOutput(nil)
	pinner := z.pin()
	defer func() {
		pinner.Unpin()
		utils.Debug("InflateEnd Unpinned")
	}()
	utils.Debug("InflateEnd %p next_int: %p out: next_out %p", z, z.strm.next_in, z.strm.next_out)

	return ZConstant(C.InflateEnd(&z.strm))
}

// SetInput sets the input buffer for the stream.
// The input buffer is not copied and must not be modified during the stream operation.
// The input buffer must have an underlying capacity larger than its size by at least one.
// This is to avoid the case where the stream ends at the end of the buffer which would
// result in an internal state that points past the end of the buffer and causes an error
// "found pointer to free object".
func (z *zstream) SetInput(in []byte) {
	z.in = in
	z.strm.avail_in = C.uint(len(in))
}

// SetOutput sets the output buffer for the stream.
// The output buffer is not copied and must not be modified during the stream operation.
// The output buffer must have an underlying capacity larger than its size by at least one.
// This is to avoid the case where the stream ends at the end of the buffer which would
// result in an internal state that points past the end of the buffer and causes an error
// "found pointer to free object".
func (z *zstream) SetOutput(out []byte) {
	z.out = out
	z.strm.avail_out = C.uint(len(out))
}

// Deflate compresses as much data as possible, and stops when the input buffer becomes empty or the output buffer becomes full.
// For more details, see http://zlib.net/manual.html#Basic
func (z *zstream) Deflate(flush ZConstant) ZConstant {
	var ret ZConstant
	z.wrapOp(func() {
		ret = ZConstant(C.Deflate(&z.strm, C.int(flush)))
	})
	return ret
}

// Inflate decompresses as much data as possible, and stops when the input buffer becomes empty or the output buffer becomes full.
// For more details, see http://zlib.net/manual.html#Basic
func (z *zstream) Inflate(flush ZConstant) ZConstant {
	var ret ZConstant
	z.wrapOp(func() {
		ret = ZConstant(C.Inflate(&z.strm, C.int(flush)))
	})
	return ret
}

// ProducedOutput returns the number of bytes produced in the output buffer.
func (z *zstream) ProducedOutput() int {
	return len(z.out) - int(z.strm.avail_out)
}

// OutputBufferIsFull returns true if the output buffer is full.
func (z *zstream) OutputBufferIsFull() bool {
	return z.strm.avail_out == 0
}

// AvailIn returns the number of bytes yet to be consumed from the input buffer.
func (z *zstream) AvailIn() int {
	return int(z.strm.avail_in)
}

// wrapOp wraps a call to Deflate or Inflate.
// In order to avoid C pointers having access to free memory in Go memory,
// next_in and next_out is reset to nil after each call to Deflate or Inflate.
// They are set again to the correct position in the buffer before each call.
// The correct position is inferred from the length of the buffer and the
// value of avail_in and avail_out.
func (z *zstream) wrapOp(f func()) {
	// Pin buffers
	pinner := z.pin()
	defer func() {
		pinner.Unpin()
		utils.Debug("wrapOp Unpinned")
	}()

	// Set C pointers
	if z.strm.avail_in == 0 {
		z.strm.next_in = nil
	} else {
		z.strm.next_in = (*C.Bytef)(&z.in[len(z.in)-int(z.strm.avail_in)])
	}
	if z.strm.avail_out == 0 {
		z.strm.next_out = nil
	} else {
		z.strm.next_out = (*C.Bytef)(&z.out[len(z.out)-int(z.strm.avail_out)])
	}

	// Call f
	utils.Debug("Before %p next_in:%v, avail_in:%v, next_out:%v, avail_out:%v", &z.strm, z.strm.next_in, z.strm.avail_in, z.strm.next_out, z.strm.avail_out)
	f()
	utils.Debug("After %p next_in:%v, avail_in:%v, next_out:%v, avail_out:%v", &z.strm, z.strm.next_in, z.strm.avail_in, z.strm.next_out, z.strm.avail_out)

	// Reset C pointers
	z.strm.next_in = nil
	z.strm.next_out = nil
	utils.Debug("Cleaned %p next_in:%v, avail_in:%v, next_out:%v, avail_out:%v", &z.strm, z.strm.next_in, z.strm.avail_in, z.strm.next_out, z.strm.avail_out)

	// Unpin buffers - deferred
}

// pin pins the zstream and its buffers to prevent the GC from moving them.
func (z *zstream) pin() runtime.Pinner {
	pinner := runtime.Pinner{}
	pinner.Pin(&z.strm)
	if len(z.in) > 0 {
		pinner.Pin(&z.in[0])
		pinner.Pin(&z.in)
	}
	if len(z.out) > 0 {
		pinner.Pin(&z.out[0])
		pinner.Pin(&z.out)
	}
	return pinner
}
