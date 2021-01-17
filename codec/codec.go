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
import (
	"errors"
	"syscall"
)

func init() {
	C.avcodec_register_all()
}

type Packet struct {
	avPacket *C.struct_AVPacket
}

//初始化Packet，用于存放编码压缩过的数据
func (packet *Packet) Init() error {
	avPacket := C.av_packet_alloc()
	if avPacket == nil {
		return errors.New(syscall.ENOMEM.Error())
	}
	C.av_init_packet((*C.struct_AVPacket)(avPacket))
	avPacket.data = nil
	avPacket.size = 0
	packet.avPacket = avPacket
	return nil
}

func (packet *Packet) Deinit() {
	C.av_packet_free(&packet.avPacket)
	packet.avPacket = nil
}
