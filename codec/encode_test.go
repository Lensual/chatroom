package codec_test

import (
	"bytes"
	"encoding/binary"
	"io/ioutil"
	"os"
	"testing"

	"github.com/Lensual/chatroom/codec"
)

func intToBytes(n int) []byte {
	data := int32(n)
	bytebuf := bytes.NewBuffer([]byte{})
	_ = binary.Write(bytebuf, binary.BigEndian, data)
	return bytebuf.Bytes()
}

func writePkt(pkt *[][]byte, file *os.File) {
	for _, v := range *pkt {
		// 先写入packet size，便于解码时读取
		pktSize := intToBytes(len(v))
		_, err := file.Write(pktSize)
		if err != nil {
			panic(err)
		}

		_, err = file.Write(v)
		if err != nil {
			panic(err)
		}
	}
}

func testEncoderOpus(t *testing.T) {
	format := codec.SampleFmt_S16
	layout := codec.ChLayout_Mono
	sampleRate := 48000

	//初始化编码器
	enc := codec.Encoder{}
	err := enc.Init("libopus", format, layout, sampleRate, 32000)
	if err != nil {
		t.Fatal(err)
	}

	//初始化帧
	frame := codec.Frame{}
	frameSize := enc.GetFrameSize()
	err = frame.InitByFormat(format, layout, frameSize)
	if err != nil {
		t.Fatal(err)
	}

	//打开文件
	data, err := ioutil.ReadFile("../test/48000_mono.pcm")
	if err != nil {
		t.Fatal(err)
	}

	out, err := os.Create("../test/test_encode_opus.data")
	if err != nil {
		t.Fatal(err)
	}

	//每次取一帧的样本
	step := frame.GetDataSize()
	for i := 0; i < len(data); i += step {
		if i+step > len(data) {
			break
		}
		sample := data[i : i+step]
		err = frame.MakeWriteable()
		if err != nil {
			t.Fatal(err)
		}

		frame.Write(&sample, step)

		pkt, err := enc.Encode(&frame)
		if err != nil {
			t.Fatal(err)
		}

		// fmt.Printf("%v\n", pkt)

		writePkt(pkt, out)
	}

	//flush
	pkt, err := enc.Encode(nil)
	if err != nil {
		t.Fatal(err)
	}
	writePkt(pkt, out)

	out.Close()
	enc.Deinit()
}
