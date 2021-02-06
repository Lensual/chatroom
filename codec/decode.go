package codec

/*
#include "libavcodec/avcodec.h"
#include "libavutil/error.h"
#include <string.h>
*/
import "C"
import (
	"errors"
	"syscall"
	"unsafe"
)

//Decoder 解码器
type Decoder struct {
	avCodec    *C.struct_AVCodec
	avCodecCtx *C.struct_AVCodecContext
	// avCodecParams *C.struct_AVCodecParameters
	packet *Packet
	frame  *Frame
}

// 初始化解码器
func (dec *Decoder) Init(decoderName string, fmt SampleFormat, layout ChannelLayout, sampleRate int) error {
	noerr := false
	defer func() {
		if !noerr {
			dec.Deinit()
		}
	}()

	//查找解码器
	codesName := C.CString(decoderName)
	avCodec := C.avcodec_find_decoder_by_name(codesName)
	C.free(unsafe.Pointer(codesName))
	if avCodec == nil {
		return errors.New("decoder is not found")
	}
	dec.avCodec = avCodec

	//初始化解码器上下文
	avCodecCtx := C.avcodec_alloc_context3(avCodec)
	if avCodecCtx == nil {
		return errors.New(syscall.ENOMEM.Error())
	}
	dec.avCodecCtx = avCodecCtx

	//配置解码器参数
	avCodecCtx.sample_fmt = C.enum_AVSampleFormat(fmt)
	avCodecCtx.sample_rate = C.int(sampleRate)
	avCodecCtx.channel_layout = C.ulonglong(layout)
	avCodecCtx.request_channel_layout = C.ulonglong(layout)
	avCodecCtx.channels = C.av_get_channel_layout_nb_channels(C.ulonglong(layout))

	//打开解码器
	code := int(C.avcodec_open2(avCodecCtx, avCodec, nil))
	if code != 0 {
		return errors.New((err2str(code)))
	}

	//初始化Packet
	packet := &Packet{}
	err := packet.Init(0)
	if err != nil {
		return err
	}
	dec.packet = packet

	//初始化Frame
	frame := &Frame{}
	err = frame.Init()
	if err != nil {
		return err
	}
	dec.frame = frame

	//返回
	noerr = true
	return nil
}

//释放解码器
func (dec *Decoder) Deinit() {
	if dec.avCodec != nil {
		// C.av_free(unsafe.Pointer(dec.avCodec)) //不能释放这个，这个是全局静态的
		dec.avCodec = nil
	}
	if dec.avCodecCtx != nil {
		C.avcodec_close(dec.avCodecCtx)
		C.avcodec_free_context(&dec.avCodecCtx)
		dec.avCodecCtx = nil
	}
	if dec.packet != nil {
		dec.packet.Unref()
		dec.packet.Deinit()
		dec.packet = nil
	}
	if dec.frame != nil {
		dec.frame.Unref()
		dec.frame.Deinit()
		dec.frame = nil
	}
}

// Decode 解码
// 返回值:output *[][]byte,eof bool, err error
func (dec *Decoder) Decode(input *[]byte) (*[][]byte, bool, error) {
	//是否为flush请求
	if input == nil {
		//flush请求
		code := int(C.avcodec_send_packet(dec.avCodecCtx, nil))
		if code < 0 {
			return nil, false, errors.New((err2str(code)))
		}
	} else {
		//正常请求

		//将输入数据转换为packet结构
		dec.packet.Parse(input)

		//发送待解码数据
		code := int(C.avcodec_send_packet(dec.avCodecCtx, dec.packet.avPacket))
		if code < 0 {
			return nil, false, errors.New(err2str(code))
		}
	}

	output := make([][]byte, 0)
	eof := false

	for {
		//接收解码后的数据
		code := int(C.avcodec_receive_frame(dec.avCodecCtx, dec.frame.avFrame))
		if code < 0 {
			if code == -C.EAGAIN {
				//可能需要更多数据解码
				break
			} else if code == C.AVERROR_EOF {
				eof = true
				break
			} else if code < 0 {
				return nil, false, errors.New(err2str(code))
			}
		}

		size := dec.GetRealFrameSize()
		sample := C.GoBytes(unsafe.Pointer(dec.frame.avFrame.data[0]), C.int(size))
		output = append(output, sample)
	}

	return &output, eof, nil
}

func (dec *Decoder) GetRealFrameSize() int {
	if dec.avCodecCtx != nil {
		nbSamples := int(dec.frame.avFrame.nb_samples)
		return nbSamples * int(C.av_get_bytes_per_sample(dec.avCodecCtx.sample_fmt))
	}
	return 0
}
