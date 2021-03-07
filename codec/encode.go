package codec

/*
#include "libavcodec/avcodec.h"
#include "libavutil/error.h"
#include <string.h>

//TODO: 封装成GO
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

//Encoder 编码器
type Encoder struct {
	avCodec         *C.struct_AVCodec
	avCodecCtx      *C.struct_AVCodecContext
	packetGenerator func() *Packet
	packetRecycler  func(*Packet)
}

// 初始化编码器
func (enc *Encoder) Init(encoderName string, fmt SampleFormat, layout ChannelLayout, sampleRate int, bitRate int) error {
	noerr := false
	defer func() {
		if !noerr {
			enc.Deinit()
		}
	}()

	//查找编码器
	codesName := C.CString(encoderName)
	avCodec := C.avcodec_find_encoder_by_name(codesName)
	C.free(unsafe.Pointer(codesName))
	if avCodec == nil {
		return errors.New(encoderName + "encoder is not found")
	}
	enc.avCodec = avCodec

	//检查参数
	if int(C.check_sample_rate(avCodec, C.int(sampleRate))) == 0 {
		return errors.New("unsupport sample rate")
	}
	if int(C.check_channel_layout(avCodec, C.uint64_t(layout))) == 0 {
		return errors.New("unsupport layouts")
	}
	if int(C.check_sample_fmt(avCodec, C.enum_AVSampleFormat(fmt))) == 0 {
		return errors.New("unsupport sample format")
	}

	//初始化编码器上下文
	avCodecCtx := C.avcodec_alloc_context3(avCodec)
	if avCodecCtx == nil {
		return errors.New(syscall.ENOMEM.Error())
	}
	enc.avCodecCtx = avCodecCtx

	//配置编码器参数
	avCodecCtx.bit_rate = C.longlong(bitRate)
	avCodecCtx.sample_fmt = C.enum_AVSampleFormat(fmt)
	avCodecCtx.sample_rate = C.int(sampleRate)
	avCodecCtx.channel_layout = C.ulonglong(layout)
	avCodecCtx.channels = C.av_get_channel_layout_nb_channels(C.ulonglong(layout))

	//打开编码器
	code := int(C.avcodec_open2(avCodecCtx, avCodec, nil))
	if code < 0 {
		return errors.New((err2str(code)))
	}

	//返回
	noerr = true
	return nil
}

//释放编码器
func (enc *Encoder) Deinit() {
	if enc.avCodec != nil {
		// C.av_free(unsafe.Pointer(enc.avCodec)) //不能释放这个，这个是全局静态的
		enc.avCodec = nil
	}
	if enc.avCodecCtx != nil {
		C.avcodec_close(enc.avCodecCtx)
		C.avcodec_free_context(&enc.avCodecCtx)
		enc.avCodecCtx = nil
	}
	enc.packetGenerator = nil
}

func (enc *Encoder) SetPacketGenerator(f func() *Packet) {
	enc.packetGenerator = f
}

func (enc *Encoder) SetPacketRecycler(f func(*Packet)) {
	enc.packetRecycler = f
}

func (enc *Encoder) SendFrame(frame *Frame) error {
	code := int(C.avcodec_send_frame(enc.avCodecCtx, frame.avFrame))
	if code < 0 {
		return errors.New((err2str(code)))
	}
	return nil
}

func (enc *Encoder) RecvPacket(packet *Packet) (bool, error) {
	code := int(C.avcodec_receive_packet(enc.avCodecCtx, packet.avPacket))
	if code == -C.EAGAIN || code == C.AVERROR_EOF {
		return true, nil
	} else if code < 0 {
		return false, errors.New((err2str(code)))
	}
	return false, nil
}

//回收Packet
func (enc *Encoder) recyclingPacket(packet *Packet) {
	packet.Unref()
	if enc.packetRecycler != nil {
		enc.packetRecycler(packet)
	} else {
		packet.Deinit()
	}
}

// 编码
// 注意：可能会返回多个Packet
// 注意：当出现错误时也可能会返回Packet，应当对返回的Packet进行释放处理
// 返回值: output *[]Packet, err error,
func (enc *Encoder) EncodeToPacketByFrame(frame *Frame) (*[]Packet, error) {
	var avFrame *C.AVFrame
	//是否为flush请求
	if frame == nil {
		//flush请求
		avFrame = nil
	} else {
		//正常请求
		avFrame = frame.avFrame
	}

	//编码
	code := int(C.avcodec_send_frame(enc.avCodecCtx, avFrame))
	if code < 0 {
		return nil, errors.New((err2str(code)))
	}

	packets := make([]Packet, 0)

	var err error

	for {
		var packet *Packet

		//如果有生成器则使用生成器
		if enc.packetGenerator != nil {
			packet = enc.packetGenerator()
		}
		if packet == nil {
			packet = &Packet{}
			err = packet.Init(0)
			if err != nil {
				break
			}
		}

		code := int(C.avcodec_receive_packet(enc.avCodecCtx, packet.avPacket))
		if code < 0 {
			//回收Packet
			enc.recyclingPacket(packet)

			if code == -C.EAGAIN || code == C.AVERROR_EOF {
				break
			} else {
				err = errors.New((err2str(code)))
				break
			}
		}
		packets = append(packets, *packet)
	}

	return &packets, err
}

// 编码
// 注意：可能会返回多个Packet
// 注意：当出现错误时也可能会返回Packet，应当对返回的Packet进行释放处理
// 返回值: output *[]Packet, err error
func (enc *Encoder) EncodeToPacketByData(data *[]byte) (*[]Packet, error) {
	var frame *Frame
	var err error
	var packets *[]Packet

	//是否为flush请求
	if data == nil {
		//flush请求
		frame = nil
	} else {
		//正常请求
		frame = &Frame{}
		err = frame.InitByFormat(SampleFormat(enc.avCodecCtx.sample_fmt),
			ChannelLayout(enc.avCodecCtx.channel_layout),
			enc.GetFrameSize())
		if err != nil {
			goto END
		}
		err = frame.MakeWriteable()
		if err != nil {
			goto END
		}
		frame.Write(data, frame.GetDataSize())
	}

	//编码
	packets, err = enc.EncodeToPacketByFrame(frame)

END:
	//TODO Frame FIFO
	if frame != nil {
		frame.Unref()
		frame.Deinit()
	}
	return packets, err
}

// 编码
// 注意：可能会返回多个Packet数据
// 注意：当出现错误时也可能会返回部分数据，但因为是GO对象所以不需要释放处理
// 返回值: output *[][]byte, err error
func (enc *Encoder) EncodeToDataByData(data *[]byte) (*[][]byte, error) {
	var packets *[]Packet
	var err error
	packets, err = enc.EncodeToPacketByData(data)

	output := make([][]byte, 0)
	for _, v := range *packets {
		outputData := v.GetData()
		output = append(output, *outputData)
		//回收Packet
		enc.recyclingPacket(&v)
	}

	return &output, err
}

//获取编码器一帧处理的样本数，失败返回0
func (enc *Encoder) GetFrameSize() int {
	if enc.avCodecCtx != nil {
		return int(enc.avCodecCtx.frame_size)
	}
	return 0
}

//获取一帧所占用的空间大小
func (enc *Encoder) GetSize() int {
	if enc.avCodecCtx != nil {
		return enc.GetFrameSize() * GetBytesPerSample(SampleFormat(enc.avCodecCtx.sample_fmt))
	}
	return 0
}

//获取ExtraData
func (enc *Encoder) GetExtraData() *[]byte {
	if enc.avCodecCtx != nil {
		data := C.GoBytes(unsafe.Pointer(enc.avCodecCtx.extradata), enc.avCodecCtx.extradata_size)
		return &data
	}
	return nil
}
