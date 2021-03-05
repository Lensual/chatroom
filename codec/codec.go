package codec

/*
#cgo CFLAGS:  -I../lib/build/include -Wno-deprecated-declarations
#cgo LDFLAGS: -L../lib/build/lib
#cgo LDFLAGS: -lavcodec
#cgo LDFLAGS: -lavutil
#cgo LDFLAGS: -lavformat
#cgo LDFLAGS: -lopus
#cgo LDFLAGS: -lavfilter
#cgo LDFLAGS: -lswresample
// #cgo LDFLAGS: -lswscale
// #cgo LDFLAGS: -lavdevice
// #cgo LDFLAGS: -lm

#cgo !windows LDFLAGS: -lm
#cgo windows LDFLAGS: -lbcrypt -lssp
#include "libavcodec/avcodec.h"
*/
import "C"
import (
	"errors"
	"syscall"
	"unsafe"
)

func init() {
	C.avcodec_register_all()
}

type Packet struct {
	avPacket *C.struct_AVPacket
}

//初始化Packet，用于存放编码压缩过的数据
func (packet *Packet) Init(bufferSize int) error {
	var avPacket *C.struct_AVPacket
	if bufferSize > 0 {
		code := int(C.av_new_packet(avPacket, C.int(bufferSize)))
		if code != 0 {
			return errors.New(err2str(code))
		}
	} else {
		avPacket = C.av_packet_alloc()
		if avPacket == nil {
			return errors.New(syscall.ENOMEM.Error())
		}
		// C.av_init_packet((*C.struct_AVPacket)(avPacket))
		avPacket.data = nil
		avPacket.size = 0
	}
	packet.avPacket = avPacket
	return nil
}

//释放Packet
func (packet *Packet) Deinit() {
	C.av_packet_free(&packet.avPacket)
	packet.avPacket = nil
}

//释放Buffer，重置Packet
func (packet *Packet) Unref() {
	C.av_packet_unref(packet.avPacket)
}

//将数据转换为Packet结构，注意，这样不使用引用计数特性
func (packet *Packet) Parse(data *[]byte) {
	cdata := C.CBytes(*data)
	packet.avPacket.data = (*C.uchar)(cdata)
	packet.avPacket.size = C.int(len(*data))
}

func (packet *Packet) GetData() *[]byte {
	pkt := C.GoBytes(unsafe.Pointer(packet.avPacket.data), packet.avPacket.size)
	return &pkt
}

//单链表 FIFO
type PacketPool struct {
	head   *packetPoolNode
	tail   *packetPoolNode
	length int
}

type packetPoolNode struct {
	//next 指向head方向
	next  *packetPoolNode
	value *Packet
}

func (pool *PacketPool) Pop() *Packet {
	//TODO lock
	if pool.tail == nil {
		return nil
	}
	remove := pool.tail

	ret := pool.tail.value
	pool.tail = pool.tail.next
	pool.length--

	remove.next = nil
	remove.value = nil
	remove = nil
	return ret
}

func (pool *PacketPool) Push(pkt *Packet) {
	if pkt != nil {
		node := &packetPoolNode{
			value: pkt,
			next:  nil,
		}
		if pool.length > 0 {
			pool.head.next = node
		} else {
			pool.tail = node
		}
		pool.head = node
		pool.length++
	}
}

func (pool *PacketPool) GetLength() int {
	return pool.length
}
