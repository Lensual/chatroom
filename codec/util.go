package codec

/*
#cgo LDFLAGS: -lavutil

#include "libavutil/audio_fifo.h"
#include "libavutil/error.h"
#include "libavutil/frame.h"
#include "libavutil/channel_layout.h"
#include "libavutil/samplefmt.h"

static const char *error2string(int code) { return av_err2str(code); }
*/
import "C"
import (
	"errors"
	"syscall"
	"unsafe"
)

/***************************
*		AudioFIFO
***************************/

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

/***************************
*		Frame
***************************/

//帧结构体 一个帧结构体可能包含多个帧
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

//带格式初始化
func (frame *Frame) InitByFormat(fmt SampleFormat, layout ChannelLayout, size int) error {
	err := frame.Init()
	if err != nil {
		return err
	}
	frame.SetFormat(fmt, layout)
	frame.SetSize(size)
	err = frame.GetBuffer()
	if err != nil {
		return err
	}
	return nil
}

//设置样本格式
func (frame *Frame) SetFormat(fmt SampleFormat, layout ChannelLayout) {
	frame.avFrame.format = C.int(fmt)
	frame.avFrame.channel_layout = C.uint64_t(layout)
}

//设置大小（包含的样本数）
func (frame *Frame) SetSize(size int) {
	frame.avFrame.nb_samples = C.int(size)
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

//分配引用计数Buffer，需要先设置好nb_samples、format、channel_layout
func (frame *Frame) GetBuffer() error {
	code := int(C.av_frame_get_buffer(frame.avFrame, 0))
	if code < 0 {
		return errors.New((err2str(code)))
	}
	return nil
}

//让Buffer变得可写，如果不可写则分配新的Buffer
func (frame *Frame) MakeWriteable() error {
	code := int(C.av_frame_make_writable(frame.avFrame))
	if code < 0 {
		return errors.New((err2str(code)))
	}
	return nil
}

//写入Frame
func (frame *Frame) Write(data *[]byte, len int) {
	cData := C.CBytes(*data)
	C.memcpy(unsafe.Pointer(frame.avFrame.data[0]), cData, C.uint64_t(len))
	C.free(cData)
}

//获取数据大小
func (frame *Frame) GetDataSize() int {
	channels := GetChannelLayoutNbChannels(ChannelLayout(frame.avFrame.channel_layout))
	sampleSize := SampleFormat(frame.avFrame.format).GetSize()
	return channels * int(frame.avFrame.nb_samples) * sampleSize
}

func (frame *Frame) GetLineSize() int {
	return int(frame.avFrame.linesize[0])
}

/***************************
*		ChannelLayout
***************************/

//通道布局
type ChannelLayout C.uint64_t

const (
	ChLayout_Mono              ChannelLayout = C.AV_CH_LAYOUT_MONO
	ChLayout_Stereo            ChannelLayout = C.AV_CH_LAYOUT_STEREO
	ChLayout_2Point1           ChannelLayout = C.AV_CH_LAYOUT_2POINT1
	ChLayout_2_1               ChannelLayout = C.AV_CH_LAYOUT_2_1
	ChLayout_Surround          ChannelLayout = C.AV_CH_LAYOUT_SURROUND
	ChLayout_3Point1           ChannelLayout = C.AV_CH_LAYOUT_3POINT1
	ChLayout_4Point0           ChannelLayout = C.AV_CH_LAYOUT_4POINT0
	ChLayout_4Point1           ChannelLayout = C.AV_CH_LAYOUT_4POINT1
	ChLayout_2_2               ChannelLayout = C.AV_CH_LAYOUT_2_2
	ChLayout_Quad              ChannelLayout = C.AV_CH_LAYOUT_QUAD
	ChLayout_5Point0           ChannelLayout = C.AV_CH_LAYOUT_5POINT0
	ChLayout_5Point1           ChannelLayout = C.AV_CH_LAYOUT_5POINT1
	ChLayout_5Point0_Back      ChannelLayout = C.AV_CH_LAYOUT_5POINT0_BACK
	ChLayout_5Point1_Back      ChannelLayout = C.AV_CH_LAYOUT_5POINT1_BACK
	ChLayout_6Point0           ChannelLayout = C.AV_CH_LAYOUT_6POINT0
	ChLayout_6Point0_Front     ChannelLayout = C.AV_CH_LAYOUT_6POINT0_FRONT
	ChLayout_Hexagonal         ChannelLayout = C.AV_CH_LAYOUT_HEXAGONAL
	ChLayout_6Point1           ChannelLayout = C.AV_CH_LAYOUT_6POINT1
	ChLayout_6Point1_Back      ChannelLayout = C.AV_CH_LAYOUT_6POINT1_BACK
	ChLayout_6Point1_Front     ChannelLayout = C.AV_CH_LAYOUT_6POINT1_FRONT
	ChLayout_7Point0           ChannelLayout = C.AV_CH_LAYOUT_7POINT0
	ChLayout_7Point0_Front     ChannelLayout = C.AV_CH_LAYOUT_7POINT0_FRONT
	ChLayout_7Point1           ChannelLayout = C.AV_CH_LAYOUT_7POINT1
	ChLayout_7Point1_Wide      ChannelLayout = C.AV_CH_LAYOUT_7POINT1_WIDE
	ChLayout_7Point1_Wide_Back ChannelLayout = C.AV_CH_LAYOUT_7POINT1_WIDE_BACK
	ChLayout_Octagonal         ChannelLayout = C.AV_CH_LAYOUT_OCTAGONAL
	ChLayout_Hexadecagonal     ChannelLayout = C.AV_CH_LAYOUT_HEXADECAGONAL
	ChLayout_Stero_Downmix     ChannelLayout = C.AV_CH_LAYOUT_STEREO_DOWNMIX
)

//获取通道布局对应的通道数
func (layout ChannelLayout) GetChannels() int {
	return GetChannelLayoutNbChannels(layout)
}

/***************************
*		SampleFormat
***************************/

//样本数据格式
type SampleFormat C.enum_AVSampleFormat

const (
	SampleFmt_None   SampleFormat = C.AV_SAMPLE_FMT_NONE
	SampleFmt_U8     SampleFormat = C.AV_SAMPLE_FMT_U8  ///< unsigned 8 bits
	SampleFmt_S16    SampleFormat = C.AV_SAMPLE_FMT_S16 ///< signed 16 bits
	SampleFmt_S32    SampleFormat = C.AV_SAMPLE_FMT_S32 ///< signed 32 bits
	SampleFmt_Float  SampleFormat = C.AV_SAMPLE_FMT_FLT ///< float
	SampleFmt_Double SampleFormat = C.AV_SAMPLE_FMT_DBL ///< double

	SampleFmt_U8P     SampleFormat = C.AV_SAMPLE_FMT_U8P  ///< unsigned 8 bits, planar
	SampleFmt_S16P    SampleFormat = C.AV_SAMPLE_FMT_S16P ///< signed 16 bits, planar
	SampleFmt_S32P    SampleFormat = C.AV_SAMPLE_FMT_S32P ///< signed 32 bits, planar
	SampleFmt_FloatP  SampleFormat = C.AV_SAMPLE_FMT_FLTP ///< float, planar
	SampleFmt_DoubleP SampleFormat = C.AV_SAMPLE_FMT_DBLP ///< double, planar
	SampleFmt_S64     SampleFormat = C.AV_SAMPLE_FMT_S64  ///< signed 64 bits
	SampleFmt_S64P    SampleFormat = C.AV_SAMPLE_FMT_S64P ///< signed 64 bits, planar

	SampleFmt_NB SampleFormat = C.AV_SAMPLE_FMT_NB ///< Number of sample formats. DO NOT USE if linking dynamically
)

//获取SampleFormat对应的字符串名称
func (format SampleFormat) GetName() string {
	return GetSampleFmtName(format)
}

//获取指定样本格式占用多少字节
func (format SampleFormat) GetSize() int {
	return GetBytesPerSample(format)
}

/***************************
*		Functions
***************************/

//获取错误字符串
func err2str(code int) string {
	cStr := C.error2string(C.int(code))
	goStr := C.GoString(cStr)
	// C.free(unsafe.Pointer(cStr))
	return goStr
}

//获取SampleFormat对应的字符串
func GetSampleFmtName(format SampleFormat) string {
	cStr := C.av_get_sample_fmt_name(C.enum_AVSampleFormat(format))
	return C.GoString(cStr)
}

//获取指定样本格式占用多少字节
func GetBytesPerSample(format SampleFormat) int {
	return int(C.av_get_bytes_per_sample(int32(format)))
}

// //依据channel，nb_sample，sample_fmt 计算缓冲器大小
func GetSamplesBufferSize(channels int, nbSamples int, sampleFmt SampleFormat) int {
	return int(C.av_samples_get_buffer_size(nil,
		C.int(channels), C.int(nbSamples), C.enum_AVSampleFormat(sampleFmt), 0))
}

//获取通道布局对应的通道数
func GetChannelLayoutNbChannels(layout ChannelLayout) int {
	return int(C.av_get_channel_layout_nb_channels(C.uint64_t(layout)))
}
