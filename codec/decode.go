package codec

/*
#cgo CFLAGS:  -I../lib/FFmpeg
#cgo LDFLAGS: -L../lib/FFmpeg/libavcodec -lavcodec -L../lib/FFmpeg/libavutil
#cgo LDFLAGS: -L../lib/FFmpeg/libavutil -lavutil
#cgo LDFLAGS: -L../lib/FFmpeg/libswresample -lswresample
#cgo LDFLAGS: -lbcrypt
#include "libavcodec/avcodec.h"
*/
import "C"

type (
	Codec           C.struct_AVCodec
	CodecContext    C.struct_AVCodecContext
	CodecParameters C.struct_AVCodecParameters
)

//OpusDecoder OPUS解码器
type OpusDecoder struct {
	codec       *C.struct_AVCodec
	codecCtx    *C.struct_AVCodecContext
	codecParams *C.struct_AVCodecParameters
	Channel     int
	BitRate     int
}

//Init 初始化OPUS解码器
func (opusDec *OpusDecoder) Init() error {
	//初始化解码器
	codec := C.avcodec_find_decoder(C.AV_CODEC_ID_OPUS)
	opusDec.codec = codec

	codecCtx := C.avcodec_alloc_context3(codec)
	opusDec.codecCtx = codecCtx

	//配置解码器参数
	params := C.avcodec_parameters_alloc()
	params.codec_type = C.AVMEDIA_TYPE_AUDIO
	params.codec_id = C.AV_CODEC_ID_OPUS
	params.channels = 1
	params.format = C.AV_SAMPLE_FMT_U8 //TODO 需要测试
	params.sample_rate = 22050
	opusDec.codecParams = params

	C.avcodec_parameters_to_context(codecCtx, params)

	//打开解码器
	C.avcodec_open2(codecCtx, codec, nil)

	//返回 TODO
	return nil
}
