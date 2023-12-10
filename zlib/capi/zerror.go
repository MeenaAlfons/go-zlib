package capi

import "fmt"

func ZError(ret ZConstant) error {
	switch ret {
	case Z_OK:
		return nil
	case Z_STREAM_END:
		return fmt.Errorf("zlib: Z_STREAM_END")
	case Z_NEED_DICT:
		return fmt.Errorf("zlib: Z_NEED_DICT")
	case Z_ERRNO:
		return fmt.Errorf("zlib: Z_ERRNO")
	case Z_STREAM_ERROR:
		return fmt.Errorf("zlib: Z_STREAM_ERROR")
	case Z_DATA_ERROR:
		return fmt.Errorf("zlib: Z_DATA_ERROR")
	case Z_MEM_ERROR:
		return fmt.Errorf("zlib: Z_MEM_ERROR")
	case Z_BUF_ERROR:
		return fmt.Errorf("zlib: Z_BUF_ERROR")
	case Z_VERSION_ERROR:
		return fmt.Errorf("zlib: Z_VERSION_ERROR")
	default:
		return fmt.Errorf("zlib: %d", ret)
	}
}
