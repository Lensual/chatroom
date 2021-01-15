package codec

/*
#include "libavcodec/avcodec.h"
#include <string.h>
*/
import "C"

//OpusDecoder OPUS解码器
type OpusDecoder struct {
	codec       *C.struct_AVCodec
	codecCtx    *C.struct_AVCodecContext
	codecParams *C.struct_AVCodecParameters
	ioCtx       *C.struct_AVIOContext
	fmtCtx      *C.struct_AVFormatContext
	ringBuf     *C.struct_RingBuffer
	audioFifo   *C.struct_AVAudioFifo
	Channel     int
	SampleRate  int
}

//Init 初始化OPUS解码器
func InitOpusDecoder(channel int, sampleRate int) (*OpusDecoder, error) {
	opusDec := OpusDecoder{}

	//查找解码器
	// codec := C.avcodec_find_decoder(C.AV_CODEC_ID_OPUS)
	codec := C.avcodec_find_decoder_by_name(C.CString("libopus"))
	opusDec.codec = codec

	//初始化解码器上下文
	codecCtx := C.avcodec_alloc_context3(codec)
	opusDec.codecCtx = codecCtx

	//配置解码器参数
	params := C.avcodec_parameters_alloc()
	params.codec_type = C.AVMEDIA_TYPE_AUDIO
	params.codec_id = C.AV_CODEC_ID_OPUS
	params.channels = C.int(channel)
	params.format = C.AV_SAMPLE_FMT_U8 //TODO 需要测试
	params.sample_rate = C.int(sampleRate)
	opusDec.codecParams = params

	C.avcodec_parameters_to_context(codecCtx, params)

	//打开解码器
	C.avcodec_open2(codecCtx, codec, nil)

	//初始化环形缓冲区
	ringBuf := initRingBuffer()
	opusDec.ringBuf = ringBuf

	//初始化IO上下文
	ioCtx := initIOContext(ringBuf)
	opusDec.ioCtx = ioCtx

	//初始化格式上下文
	fmtCtx := initFormatContext(ioCtx)
	opusDec.fmtCtx = fmtCtx

	//初始化fifo
	audioFifo := initAudioFifo(C.enum_AVSampleFormat(params.format), params.channels)
	opusDec.audioFifo = audioFifo

	//返回
	//TODO 错误处理 返回 defer
	return &opusDec, nil

}

func DeinitOpusDecoder(opusDec OpusDecoder) {
	//TODO
	// avcodec_free_context(&avctx);
}
