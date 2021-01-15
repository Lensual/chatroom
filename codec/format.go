package codec

/*
#cgo LDFLAGS: -lavutil

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
import "unsafe"

//initRingBuffer 初始化环形缓冲区 用于go给avio传递数据
func initRingBuffer() *C.struct_RingBuffer {
	ringBuf := C.malloc(C.sizeof_RingBuffer)
	return (*C.struct_RingBuffer)(ringBuf)
}

func deinitRingBuffer(ringBuf *C.struct_RingBuffer) {
	C.free((unsafe.Pointer)(ringBuf))
}

//initIOContext 初始化IO上下文
func initIOContext(ringBuf *C.struct_RingBuffer) *C.struct_AVIOContext {
	//初始化buffer ffmpeg用的
	ioCtxBuffer := C.av_malloc((C.IO_CTX_BUFFER_SIZE))

	//第四个参数是用户数据指针，暂时设置为NULL
	//第五个参数read_packet 是自己实现的io方法
	readFunc := C.readFunc(C.read_packet)
	ioCtx := C.avio_alloc_context((*C.uchar)(ioCtxBuffer),
		C.IO_CTX_BUFFER_SIZE,
		0,
		unsafe.Pointer(ringBuf),
		readFunc,
		nil,
		nil)
	return ioCtx
}

func deinitIOContext() {
	//TODO 需要释放ioCtxBuffer
	//TODO 需要始放ioCtx
	// avio_closep(&(*output_format_context)->pb);
}

//initFormatContext 初始化Format上下文
func initFormatContext(ioCtx *C.struct_AVIOContext) *C.struct_AVFormatContext {
	fmtCtx := C.avformat_alloc_context()
	fmtCtx.pb = ioCtx
	fmtCtx.flags |= C.AVFMT_FLAG_CUSTOM_IO
	return fmtCtx
}

func deinitFormatContext() {
	//TODO 需要释放fmtCtx
	// avformat_free_context(output_format_context);
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

//initAudioFifo 初始化音频FIFO，用于暂存解码后的音频样本
func initAudioFifo(sampleFmt C.enum_AVSampleFormat, channels C.int) *C.struct_AVAudioFifo {
	return C.av_audio_fifo_alloc(sampleFmt, channels, channels)
}

func deinitAudioFifo(fifo *C.struct_AVAudioFifo) {
	C.av_audio_fifo_free(fifo)
}
