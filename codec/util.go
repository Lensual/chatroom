package codec

/*
#cgo LDFLAGS: -lavutil

#include "libavutil/audio_fifo.h"
#include "libavutil/error.h"
#include "libavutil/frame.h"
#include "libavutil/channel_layout.h"

static const char *error2string(int code) { return av_err2str(code); }
*/
import "C"
import (
	"errors"
	"syscall"
)

type AudioFIFO struct {
	avAudioFIFO *C.struct_AVAudioFifo
}

// 初始化音频FIFO，用于暂存解码后的音频样本
func (fifo *AudioFIFO) Init(sampleFmt C.enum_AVSampleFormat, channels C.int) error {
	avAudioFIFO := C.av_audio_fifo_alloc(sampleFmt, channels, channels)
	if avAudioFIFO == nil {
		return errors.New(syscall.ENOMEM.Error())
	}
	fifo.avAudioFIFO = avAudioFIFO
	return nil
}

func (fifo *AudioFIFO) Deinit() {
	C.av_audio_fifo_free(fifo.avAudioFIFO)
	fifo.avAudioFIFO = nil
}

func err2str(code int) string {
	cStr := C.error2string(C.int(code))
	goStr := C.GoString(cStr)
	// C.free(unsafe.Pointer(cStr))
	return goStr
}

type Frame struct {
	avFrame *C.struct_AVFrame
}

// 初始化Frame
func (frame *Frame) Init() error {
	avFrame := C.av_frame_alloc()
	if avFrame == nil {
		return errors.New(syscall.ENOMEM.Error())
	}
	frame.avFrame = avFrame
	return nil
}

//释放Frame
func (frame *Frame) Deinit() {
	C.av_frame_free(&frame.avFrame)
	frame.avFrame = nil
}

//释放Buffer，重置Frame
func (frame *Frame) Unref() {
	C.av_frame_unref(frame.avFrame)
}

//分配Buffer，需要先设置哈nb_samples、format、channel_layout
func (frame *Frame) GetBuffer() error {
	code := int(C.av_frame_get_buffer(frame.avFrame, 0))
	if code < 0 {
		return errors.New((err2str(code)))
	}
	return nil
}

//保证Buffer可写，如果不可写则分配新的Buffer
func (frame *Frame) MakeWriteable() error {
	code := int(C.av_frame_make_writable(frame.avFrame))
	if code < 0 {
		return errors.New((err2str(code)))
	}
	return nil
}

type ChannelLayout C.uint64_t

const (
	Mono   ChannelLayout = C.AV_CH_LAYOUT_MONO
	Stereo ChannelLayout = C.AV_CH_LAYOUT_STEREO
)
