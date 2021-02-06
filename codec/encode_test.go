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
	enc := codec.Encoder{}
	err := enc.Init("libopus", codec.SampleFormatS16, codec.Mono, 48000, 32000)
	if err != nil {
		t.Fatal(err)
	}

	data, err := ioutil.ReadFile("../test/48000_mono.pcm")
	if err != nil {
		t.Fatal(err)
	}

	out, err := os.Create("../test/test_encode_opus.data")
	if err != nil {
		t.Fatal(err)
	}

	step := enc.GetRealFrameSize()
	for i := 0; i < len(data); i += step {
		if i+step > len(data) {
			break
		}
		frame := data[i : i+step]
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
