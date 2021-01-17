package codec

/*
#cgo LDFLAGS: -lavutil -lavcodec

#include "libavformat/avformat.h"
#include "libavformat/avio.h"
#include "libavutil/mem.h"
#include "libavutil/audio_fifo.h"

#include <stdlib.h>
#include <string.h>
#include <stdint.h>

#define IO_CTX_BUFFER_SIZE 16384

typedef int (*readFunc)(void *opaque, uint8_t *buf, int buf_size);
typedef int (*writeFunc)(void *opaque, uint8_t *buf, int buf_size);

//环形缓冲区
typedef struct _RingBuffer
{
    uint8_t buffer[16384];    //实际缓冲区
    uint8_t *head; // 指向环形缓冲区中 `最新数据的位置`
    uint8_t *tail; // 指向环形缓冲区中 `还没被io上下文消耗的位置`
    size_t size;   // 环形缓冲区大小
} RingBuffer;

//求可用缓冲区
int ringbuffer_get_free(RingBuffer *rb)
{
    int free;
    free = rb->head - rb->tail;
    if (free > 0)
    {
        free = rb->size - free;
    }
    else if (free < 0)
    {
        free = abs(free);
    }
    else
    {
        free = rb->size;
    }
    return free;
}

//求已用缓冲区
int ringbuffer_get_used(RingBuffer *rb)
{
    int used;
    used = rb->head - rb->tail;
    if (used < 0)
    {
        used = rb->size + used;
    }
    return used;
}

//ffmpeg自定义io的read回调
int read_packet(void *opaque, uint8_t *buf, int buf_size)
{
    RingBuffer *rb = (RingBuffer *)opaque;

    int used = ringbuffer_get_used(rb);
    buf_size = used > buf_size ? buf_size : used; //看看需要的数据量和已有的数据量那个少，按少的返回

    memcpy(buf, rb->tail, buf_size);
    rb->tail += buf_size;
    return buf_size;
}

//再自定义个write方法，搞不好会写入丢失，就先不给ffmpeg用吧 :P
int write_packet(void *opaque, uint8_t *buf, int buf_size)
{
    RingBuffer *rb = (RingBuffer *)opaque;

    int free = ringbuffer_get_free(rb);
    buf_size = free > buf_size ? buf_size : free; //看看剩余的缓冲区还够不够写，不够就只写一部分

    memcpy(rb->head, buf, buf_size);
    rb->head += buf_size;
    return buf_size;
}


*/
import "C"
import (
	"errors"
	"syscall"
	"unsafe"
)

//initRingBuffer 初始化环形缓冲区 用于go给avio传递数据
func initRingBuffer() *C.struct_RingBuffer {
	ringBuf := C.malloc(C.sizeof_RingBuffer)
	return (*C.struct_RingBuffer)(ringBuf)
}

func deinitRingBuffer(ringBuf *C.struct_RingBuffer) {
	C.free((unsafe.Pointer)(ringBuf))
}

//IO上下文
type IOContext struct {
	avIOCtx *C.struct_AVIOContext
}

//初始化IO上下文
func (ioCtx *IOContext) Init(ringBuf *C.struct_RingBuffer) error {
	if ringBuf == nil {
		return errors.New("ringBuf is nil")
	}

	//初始化buffer ffmpeg用的
	avIOCtxBuffer := C.av_malloc((C.IO_CTX_BUFFER_SIZE))
	if avIOCtxBuffer == nil {
		return errors.New(syscall.ENOMEM.Error())
	}

	//第四个参数是用户数据指针，暂时设置为NULL
	//第五个参数read_packet 是自己实现的io方法
	readFunc := C.readFunc(C.read_packet)
	avIOCtx := C.avio_alloc_context((*C.uchar)(avIOCtxBuffer),
		C.IO_CTX_BUFFER_SIZE,
		0,
		unsafe.Pointer(ringBuf),
		readFunc,
		nil,
		nil)
	if ioCtx == nil {
		C.av_free(avIOCtxBuffer)
		return errors.New(syscall.ENOMEM.Error())
	}
	ioCtx.avIOCtx = avIOCtx
	return nil
}

//释放IO上下文
func (ioCtx *IOContext) Deinit() {
	C.avio_close(ioCtx.avIOCtx) //包含了IOContext和buffer的释放，不需要重复释放
}

//Format上下文
type FormatContext struct {
	avFmtCtx *C.struct_AVFormatContext
}

// 初始化Format上下文
func (fmtCtx *FormatContext) Init(ioCtx *IOContext) error {
	if ioCtx == nil {
		return errors.New("ioCtx is nil")
	}
	avFmtCtx := C.avformat_alloc_context()
	if avFmtCtx == nil {
		return errors.New(syscall.ENOMEM.Error())
	}
	avFmtCtx.pb = ioCtx.avIOCtx
	avFmtCtx.flags |= C.AVFMT_FLAG_CUSTOM_IO
	fmtCtx.avFmtCtx = avFmtCtx
	return nil
}

//作为输入打开Format上下文
func (fmtCtx *FormatContext) Open() error {
	code := int(C.avformat_open_input(&fmtCtx.avFmtCtx, nil, nil, nil))
	if code != 0 {
		return errors.New(err2str(code))
	}
	return nil
}

//释放Format上下文
func (fmtCtx *FormatContext) Deinit() {
	C.avformat_close_input(&fmtCtx.avFmtCtx) //此处调用了avio_close释放IOContext，同时它的buffer也被释放，不需要重复释放
}

//writePacket 写入RingBuffer缓冲区，用于向ffmpeg avio提供数据
func writePacket(ringBuf unsafe.Pointer, data *[]byte) {
	length := len(*data)
	done := 0
	dataC := C.CBytes(*data)
	for {
		doneC := C.write_packet(ringBuf,
			(*C.uint8_t)(dataC),
			C.int(length))

		done += int(doneC)
		if done < length {
			//数据没写完，需要接着写
			dataC = unsafe.Pointer(uintptr(dataC) + uintptr(done))
		} else {
			break
		}
	}
	C.free(dataC)
}
