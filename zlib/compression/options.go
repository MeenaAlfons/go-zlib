package compression

import "github.com/MeenaAlfons/go-zlib/zlib/common"

type options interface {
	WindowBits() int
	Header() common.HeaderType
}

func zWindowBits(opts options) int {
	windowBits := opts.WindowBits()
	if opts.Header() == common.HeaderTypeRaw {
		windowBits = -windowBits
	}
	return windowBits
}
