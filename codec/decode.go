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
	avCodec     *C.struct_AVCodec
	avCodecCtx  *C.struct_AVCodecContext
	UsePool     bool
	packetsPool *Pool
	framesPool  *Pool
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

	//初始化对象池
	dec.framesPool = &Pool{}
	dec.packetsPool = &Pool{}

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

	//清理对象池
	if dec.framesPool != nil {
		for {
			obj := dec.framesPool.Pop()
			if obj == nil {
				break
			}
			frame := obj.(*Frame)
			frame.Unref()
			frame.Deinit()
		}
	}
	if dec.packetsPool != nil {
		for {
			obj := dec.packetsPool.Pop()
			if obj == nil {
				break
			}
			packet := obj.(*Frame)
			packet.Unref()
			packet.Deinit()
		}
	}
}

//回收Frame
func (dec *Decoder) recyclingFrame(frame *Frame) {
	frame.Unref()
	if dec.UsePool {
		dec.framesPool.Push(frame)
	} else {
		frame.Deinit()
	}
}

// 解码
// 注意：可能会返回多个Frame
// 注意：当出现错误时也可能会返回Frame，应当对返回的Frame进行释放处理
// 返回值: output []Frame, err error
func (dec *Decoder) DecodeToFrameByPacket(packet *Packet) ([]Frame, error) {
	var avPacket *C.AVPacket
	//是否为flush请求
	if packet == nil {
		//flush请求
		avPacket = nil
	} else {
		//正常请求
		avPacket = packet.avPacket
	}

	//解码
	code := int(C.avcodec_send_packet(dec.avCodecCtx, avPacket))
	if code < 0 {
		return nil, errors.New((err2str(code)))
	}

	frames := make([]Frame, 0)

	var err error

	for {
		var frame *Frame

		//如果有生成器则使用生成器
		if dec.UsePool {
			obj := dec.framesPool.Pop()
			if obj != nil {
				frame = obj.(*Frame)
			}
		}
		if frame == nil {
			frame = &Frame{}
			err = frame.Init()
			if err != nil {
				break
			}
		}

		code := int(C.avcodec_receive_frame(dec.avCodecCtx, frame.avFrame))
		if code < 0 {
			//回收Frame
			dec.recyclingFrame(frame)

			if code == -C.EAGAIN || code == C.AVERROR_EOF {
				break
			} else {
				err = errors.New(err2str(code))
				break
			}
		}

		frames = append(frames, *frame)
	}

	return frames, err
}

// 解码
// 注意：可能会返回多个Frame
// 注意：当出现错误时也可能会返回Frame，应当对返回的Frame进行释放处理
// 返回值: output []Frame, err error
func (dec *Decoder) DecodeToFrameByData(data []byte) ([]Frame, error) {
	var packet *Packet
	var err error
	var frames []Frame

	//是否为flush请求
	if data == nil {
		//flush请求
		packet = nil
	} else {
		//正常请求
		packet = &Packet{}
		err = packet.Init(0)
		if err != nil {
			goto END
		}

		//将输入数据转换为packet结构
		packet.Parse(data)
	}

	//解码
	frames, err = dec.DecodeToFrameByPacket(packet)

END:
	//TODO Packet FIFO
	if packet != nil {
		packet.Unref()
		packet.Deinit()
	}
	return frames, err
}

// 解码
// 注意：可能会返回多个Frame数据
// 注意：当出现错误时也可能会返回部分数据，应当对返回的Frame进行释放处理
// 返回值: output [][]byte, err error
func (dec *Decoder) DecodeToDataByData(data []byte) ([][]byte, error) {
	var frames []Frame
	var err error
	frames, err = dec.DecodeToFrameByData(data)

	output := make([][]byte, 0)

	for _, v := range frames {
		outputData := v.GetData()
		output = append(output, outputData)
		//回收Frame
		dec.recyclingFrame(&v)
	}

	return output, err
}
