package codec

/*
#include "libavcodec/avcodec.h"
#include "libavformat/avformat.h"
#include "libavutil/error.h"
#include <string.h>
*/
import "C"
import (
	"errors"
	"syscall"
	"unsafe"
)

//OpusDecoder OPUS解码器
type OpusDecoder struct {
	avCodec       *C.struct_AVCodec
	avCodecCtx    *C.struct_AVCodecContext
	avCodecParams *C.struct_AVCodecParameters
	ioCtx         *IOContext
	fmtCtx        *FormatContext
	ringBuf       *C.struct_RingBuffer
	audioFIFO     *AudioFIFO
	Channel       int
}

// 初始化OPUS解码器
func (opusDec *OpusDecoder) Init(channel int) error {
	noerr := false
	defer func() {
		if !noerr {
			opusDec.Deinit()
		}
	}()

	//查找解码器
	codesName := C.CString("libopus")
	avCodec := C.avcodec_find_decoder_by_name(codesName)
	C.free(unsafe.Pointer(codesName))
	if avCodec == nil {
		return errors.New("opus decoder is not found")
	}
	opusDec.avCodec = avCodec

	//初始化解码器上下文
	avCodecCtx := C.avcodec_alloc_context3(avCodec)
	if avCodecCtx == nil {
		return errors.New(syscall.ENOMEM.Error())
	}
	opusDec.avCodecCtx = avCodecCtx

	//配置解码器参数
	params := C.avcodec_parameters_alloc()
	if params == nil {
		return errors.New(syscall.ENOMEM.Error())
	}
	params.codec_type = C.AVMEDIA_TYPE_AUDIO
	params.codec_id = C.AV_CODEC_ID_OPUS
	params.channels = C.int(channel)
	params.format = C.AV_SAMPLE_FMT_U8 //TODO 需要测试
	opusDec.avCodecParams = params
	code := int(C.avcodec_parameters_to_context(avCodecCtx, params))
	if code != 0 {
		return errors.New((err2str(code)))
	}

	//打开解码器
	code = int(C.avcodec_open2(avCodecCtx, avCodec, nil))
	if code != 0 {
		return errors.New((err2str(code)))
	}

	//初始化环形缓冲区
	ringBuf := initRingBuffer()
	opusDec.ringBuf = ringBuf

	//初始化IO上下文
	ioCtx := &IOContext{}
	opusDec.ioCtx = ioCtx
	err := ioCtx.Init(ringBuf)
	if err != nil {
		return err
	}

	//初始化Format上下文
	fmtCtx := &FormatContext{}
	opusDec.fmtCtx = fmtCtx
	err = fmtCtx.Init(ioCtx)
	if err != nil {
		return err
	}
	err = fmtCtx.Open()
	if err != nil {
		return err
	}

	//初始化FIFO
	audioFIFO := &AudioFIFO{}
	opusDec.audioFIFO = audioFIFO
	err = audioFIFO.Init(C.enum_AVSampleFormat(params.format), params.channels)
	if err != nil {
		return err
	}

	//返回
	noerr = true
	return nil
}

//释放OPUS解码器
func (opusDec *OpusDecoder) Deinit() {
	if opusDec.avCodec != nil {
		// C.av_free(unsafe.Pointer(opusDec.avCodec)) //不能释放这个，这个是全局静态的
		opusDec.avCodec = nil
	}
	if opusDec.avCodecCtx != nil {
		C.avcodec_close(opusDec.avCodecCtx)
		C.avcodec_free_context(&opusDec.avCodecCtx)
		opusDec.avCodecCtx = nil
	}
	if opusDec.avCodecParams != nil {
		C.avcodec_parameters_free(&opusDec.avCodecParams)
		opusDec.avCodecParams = nil
	}
	if opusDec.ringBuf != nil {
		deinitRingBuffer(opusDec.ringBuf)
		opusDec.ringBuf = nil
	}
	if opusDec.ioCtx != nil {
		opusDec.ioCtx.Deinit()
		opusDec.ioCtx = nil
	}
	if opusDec.fmtCtx != nil {
		opusDec.fmtCtx.Deinit()
		opusDec.fmtCtx = nil
	}
	if opusDec.audioFIFO != nil {
		opusDec.audioFIFO.Deinit()
		opusDec.audioFIFO = nil
	}
}

// DecodeFrame 解码
// 返回值: err error, eof bool
func (opusDec *OpusDecoder) DecodeFrame(frame *Frame) (error, bool) {
	noerr := false
	var packet *Packet
	defer func() {
		if !noerr {
			if packet.avPacket != nil {
				packet.Unref()
				packet.Deinit()
			}
			packet = nil
			if frame.avFrame == nil {
				frame.Unref()
				frame.Deinit()
			}
			frame = nil
		}
	}()

	//初始化Packet
	packet = &Packet{}
	err := packet.Init(0)
	if err != nil {
		return err, false
	}

	//初始化Frame
	frame = &Frame{}
	err = frame.Init()
	if err != nil {
		return err, false
	}

	//从Format上下文中读出未解码的数据
	code := int(C.av_read_frame(opusDec.fmtCtx.avFmtCtx, packet.avPacket))
	if code != 0 {
		if code == C.AVERROR_EOF {
			return nil, true
		}
		return errors.New(err2str(code)), false
	}

	//发送待解码数据
	code = int(C.avcodec_send_packet(opusDec.avCodecCtx, packet.avPacket))
	if code != 0 {
		return errors.New(err2str(code)), false
	}

	//接收解码后的数据
	code = int(C.avcodec_receive_frame(opusDec.avCodecCtx, frame.avFrame))
	if code != 0 {
		if code == C.EAGAIN {
			//需要更多数据解码
			return nil, false
		} else if code == C.AVERROR_EOF {
			return nil, true
		} else {
			return errors.New(err2str(code)), false
		}
	}
	noerr = true

	//clean up
	packet.Unref()
	packet.Deinit()

	return nil, false

}
