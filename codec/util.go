package codec

/*
#cgo LDFLAGS: -lavutil

#include "libavutil/audio_fifo.h"
#include "libavutil/error.h"
#include "libavutil/frame.h"

static const char *error2string(int code) { return av_err2str(code); }
*/
import "C"
import (
	"errors"
	"syscall"
	"unsafe"
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
	C.free(unsafe.Pointer(cStr))
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

func (frame *Frame) Deinit() {
	C.av_frame_free(&frame.avFrame)
	frame.avFrame = nil
}
