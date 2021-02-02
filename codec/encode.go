package codec

/*
#include "libavcodec/avcodec.h"
#include "libavformat/avformat.h"
#include "libavutil/error.h"
#include "libavutil/channel_layout.h"
#include <string.h>

//确认样本格式是否支持
int check_sample_fmt(const AVCodec *codec, enum AVSampleFormat sample_fmt)
{
    const enum AVSampleFormat *p = codec->sample_fmts;

    while (*p != AV_SAMPLE_FMT_NONE) {
        //  printf("support sample format: %d\n", *p);
        if (*p == sample_fmt)
            return 1;
        p++;
    }
    return 0;
}

//确认采样率是否支持
int check_sample_rate(const AVCodec *codec,const int sample_rate)
{
    const int *p;

    p = codec->supported_samplerates;
    while (*p) {
        // printf("support sample rate: %d\n", *p);

		if (sample_rate == *p) {
			return 1;
		}
		p++;
	}
	return 0;
}

//确认音频布局是否支持
int check_channel_layout(const AVCodec *codec, const uint64_t ch_layout )
{
    const uint64_t *p;

    if (!codec->channel_layouts)
        return AV_CH_LAYOUT_MONO;

    p = codec->channel_layouts;
    while (*p) {
		if (ch_layout == *p) {
			return 1;
		}
        p++;
    }
    return 0;
}


*/
import "C"
import (
	"errors"
	"syscall"
	"unsafe"
)

//OpusEncoder OPUS编码器
type OpusEncoder struct {
	avCodec    *C.struct_AVCodec
	avCodecCtx *C.struct_AVCodecContext
	// avCodecParams *C.struct_AVCodecParameters
	packet     *Packet
	frame      *Frame
	Channel    int
	SampleRate int
}

// 初始化OPUS编码器
func (opusEnc *OpusEncoder) Init(layout ChannelLayout, sampleRate int, bitRate int) error {
	noerr := false
	defer func() {
		if !noerr {
			opusEnc.Deinit()
		}
	}()

	//查找编码器
	codesName := C.CString("libopus")
	avCodec := C.avcodec_find_encoder_by_name(codesName)
	C.free(unsafe.Pointer(codesName))
	if avCodec == nil {
		return errors.New("opus encoder is not found")
	}
	opusEnc.avCodec = avCodec

	//检查参数
	if int(C.check_sample_rate(avCodec, C.int(sampleRate))) == 0 {
		return errors.New("unsupport sample rate")
	}
	if int(C.check_channel_layout(avCodec, C.uint64_t(layout))) == 0 {
		return errors.New("unsupport layouts")
	}
	if int(C.check_sample_fmt(avCodec, C.AV_SAMPLE_FMT_S16)) == 0 { //TODO 这里写死 后期修正
		return errors.New("unsupport sample format")
	}

	//初始化编码器上下文
	avCodecCtx := C.avcodec_alloc_context3(avCodec)
	if avCodecCtx == nil {
		return errors.New(syscall.ENOMEM.Error())
	}
	opusEnc.avCodecCtx = avCodecCtx

	//配置编码器参数
	// params := C.avcodec_parameters_alloc()
	// if params == nil {
	// 	return errors.New(syscall.ENOMEM.Error())
	// }
	// params.codec_type = C.AVMEDIA_TYPE_AUDIO
	// params.codec_id = C.AV_CODEC_ID_OPUS
	// params.channels = C.av_get_channel_layout_nb_channels(C.uint64_t(layout))
	// params.format = C.AV_SAMPLE_FMT_S16 //TODO 需要测试
	// params.sample_rate = C.int(sampleRate)
	// params.bit_rate = C.longlong(bitRate)
	// opusEnc.avCodecParams = params
	// code := int(C.avcodec_parameters_to_context(avCodecCtx, params))
	// if code != 0 {
	// 	return errors.New((err2str(code)))
	// }

	avCodecCtx.bit_rate = C.longlong(bitRate)
	avCodecCtx.sample_fmt = C.AV_SAMPLE_FMT_S16
	avCodecCtx.sample_rate = C.int(sampleRate)
	avCodecCtx.channel_layout = C.uint64_t(layout)
	avCodecCtx.channels = C.av_get_channel_layout_nb_channels(C.uint64_t(layout))

	//打开编码器
	code := int(C.avcodec_open2(avCodecCtx, avCodec, nil))
	if code < 0 {
		return errors.New((err2str(code)))
	}

	//初始化Packet
	packet := &Packet{}
	err := packet.Init(0)
	if err != nil {
		return err
	}
	opusEnc.packet = packet

	//初始化Frame
	frame := &Frame{}
	err = frame.Init()
	if err != nil {
		return err
	}
	frame.avFrame.nb_samples = avCodecCtx.frame_size
	frame.avFrame.format = C.int(avCodecCtx.sample_fmt)
	frame.avFrame.channel_layout = avCodecCtx.channel_layout
	opusEnc.frame = frame

	//初始化Frame的buffer
	err = frame.GetBuffer()
	if err != nil {
		return err
	}

	//返回
	noerr = true
	return nil
}

//释放OPUS编码器
func (opusEnc *OpusEncoder) Deinit() {
	if opusEnc.avCodec != nil {
		// C.av_free(unsafe.Pointer(opusEnc.avCodec)) //不能释放这个，这个是全局静态的
		opusEnc.avCodec = nil
	}
	if opusEnc.avCodecCtx != nil {
		C.avcodec_close(opusEnc.avCodecCtx)
		C.avcodec_free_context(&opusEnc.avCodecCtx)
		opusEnc.avCodecCtx = nil
	}
	// if opusEnc.avCodecParams != nil {
	// 	C.avcodec_parameters_free(&opusEnc.avCodecParams)
	// 	opusEnc.avCodecParams = nil
	// }
	if opusEnc.packet != nil {
		opusEnc.packet.Unref()
		opusEnc.packet.Deinit()
		opusEnc.packet = nil
	}
	if opusEnc.frame != nil {
		opusEnc.frame.Unref()
		opusEnc.frame.Deinit()
		opusEnc.frame = nil
	}
}

// Encode编码，可能返回包含多个packet
// 返回值: err error, output *[][]byte
func (opusEnc *OpusEncoder) Encode(input *[]byte) (error, *[][]byte) {
	//检查帧大小
	if len(*input) != int(opusEnc.avCodecCtx.frame_size) {
		return errors.New("frame size dismatch"), nil
	}

	//保证帧可写
	err := opusEnc.frame.MakeWriteable()
	if err != nil {
		return err, nil
	}

	//复制到buffer
	in := C.CBytes(*input)
	C.memcpy(unsafe.Pointer(opusEnc.frame.avFrame.data[0]), in, C.size_t(opusEnc.avCodecCtx.frame_size))
	C.free(in)

	//编码
	code := int(C.avcodec_send_frame(opusEnc.avCodecCtx, opusEnc.frame.avFrame))
	if code < 0 {
		return errors.New((err2str(code))), nil
	}

	output := make([][]byte, 0)

	for {
		code = int(C.avcodec_receive_packet(opusEnc.avCodecCtx, opusEnc.packet.avPacket))
		if code == -C.EAGAIN || code == C.AVERROR_EOF {
			break
		} else if code < 0 {
			return errors.New((err2str(code))), nil
		}

		pkt := C.GoBytes(unsafe.Pointer(opusEnc.packet.avPacket.data), opusEnc.packet.avPacket.size)
		output = append(output, pkt)

		opusEnc.packet.Unref()
	}

	return nil, &output
}

//获取帧大小，失败返回0
func (opusEnc *OpusEncoder) GetFrameSize() int {
	if opusEnc.avCodecCtx != nil {
		return int(opusEnc.avCodecCtx.frame_size)
	}
	return 0
}
