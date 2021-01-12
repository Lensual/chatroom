package codec

/*
#cgo CFLAGS:  -I../lib/FFmpeg
#cgo LDFLAGS: -L../lib/FFmpeg/libavcodec -lavcodec -L../lib/FFmpeg/libavutil
#cgo LDFLAGS: -L../lib/FFmpeg/libavutil -lavutil
#cgo LDFLAGS: -L../lib/FFmpeg/libswresample -lswresample
#cgo LDFLAGS: -lbcrypt
#include "libavcodec/avcodec.h"
*/
import "C"

func init() {
	C.avcodec_register_all()
}
