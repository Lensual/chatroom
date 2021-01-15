package codec

/*
#cgo CFLAGS:  -I../lib/build/include -Wno-deprecated-declarations
#cgo LDFLAGS: -L../lib/build/lib
#cgo LDFLAGS: -lavcodec
#cgo LDFLAGS: -lavutil
#cgo LDFLAGS: -lavformat
#cgo LDFLAGS: -lopus
#cgo !windows LDFLAGS: -lm
#cgo windows LDFLAGS: -lbcrypt -lssp
#include "libavcodec/avcodec.h"
*/
import "C"

func init() {
	C.avcodec_register_all()
}
