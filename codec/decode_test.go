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

func writeFrames(frames *[][]byte, file *os.File) {
	for _, v := range *frames {
		_, err := file.Write(v)
		if err != nil {
			panic(err)
		}
	}
}

func testDecoderOpus(t *testing.T) {
	dec := codec.Decoder{}
	err := dec.Init("libopus", codec.SampleFmt_S16, codec.ChLayout_Mono, 48000)
	if err != nil {
		t.Fatal(err)
	}

	data, err := ioutil.ReadFile("../test/test_encode_opus.data")
	if err != nil {
		t.Fatal(err)
	}

	out, err := os.Create("../test/test_decode_opus.pcm")
	if err != nil {
		t.Fatal(err)
	}

	var i = 0
	for {
		if i+4 >= len(data) {
			break
		}

		pktSize := bytesToInt(data[i : i+4])
		i += 4

		pkt := data[i : i+pktSize]
		i += pktSize

		frames, _, err := dec.Decode(&pkt)
		if err != nil {
			t.Fatal(err)
		}

		writeFrames(frames, out)
	}

	//flush
	frames, _, err := dec.Decode(nil)
	if err != nil {
		t.Fatal(err)
	}
	writeFrames(frames, out)

	out.Close()
	dec.Deinit()
}
