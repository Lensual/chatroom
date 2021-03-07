package codec_test

import (
	"bytes"
	"encoding/binary"
	"io/ioutil"
	"os"
	"testing"

	"github.com/Lensual/chatroom/codec"
)

func bytesToInt(b []byte) int {
	bytebuff := bytes.NewBuffer(b)
	var data int32
	_ = binary.Read(bytebuff, binary.BigEndian, &data)
	return int(data)
}

func writeFrames(frames [][]byte, file *os.File) {
	for _, v := range frames {
		_, err := file.Write(v)
		if err != nil {
			panic(err)
		}
	}
}

func testDecoderOpus(t *testing.T) {
	//初始化解码器
	dec := codec.Decoder{
		UsePool: true, //使用对象池
	}
	err := dec.Init("libopus", codec.SampleFmt_S16, codec.ChLayout_Mono, 48000)
	if err != nil {
		t.Fatal(err)
	}

	//打开文件
	data, err := ioutil.ReadFile("../test/test_encode_opus.data")
	if err != nil {
		t.Fatal(err)
	}

	out, err := os.Create("../test/test_decode_opus.pcm")
	if err != nil {
		t.Fatal(err)
	}

	var headSize = 4
	var i = 0
	for {
		if i+headSize >= len(data) {
			break
		}

		//从head获取大小
		pktSize := bytesToInt(data[i : i+headSize])
		i += headSize

		//获取packet
		packet := data[i : i+pktSize]
		i += pktSize

		frames, err := dec.DecodeToDataByData(packet)
		if err != nil {
			t.Fatal(err)
		}

		writeFrames(frames, out)
	}

	//flush
	frames, err := dec.DecodeToDataByData(nil)
	if err != nil {
		t.Fatal(err)
	}
	writeFrames(frames, out)

	out.Close()
	dec.Deinit()
}
