package codec_test

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/Lensual/chatroom/codec"
)

func writePkt(pkt *[][]byte, file *os.File) {
	for _, v := range *pkt {
		_, err := file.Write(v)
		if err != nil {
			panic(err)
		}
	}
}

func TestOpusEncoder(t *testing.T) {
	enc := codec.OpusEncoder{}
	err := enc.Init(codec.SampleFormatS16, codec.Mono, 16000, 64000)
	if err != nil {
		t.Fatal(err)
	}

	data, err := ioutil.ReadFile("../test/test.pcm")
	if err != nil {
		t.Fatal(err)
	}

	out, err := os.Create("../test/test.opus")
	if err != nil {
		t.Fatal(err)
	}

	step := enc.GetFrameSize() //帧大小为步进
	for i := 0; i < len(data); i += step {
		if i+step > len(data) {
			break
		}
		frame := data[i : i+step]
		err, pkt := enc.Encode(&frame)
		if err != nil {
			t.Fatal(err)
		}

		// fmt.Printf("%v\n", pkt)

		writePkt(pkt, out)

	}

	//flush
	err, pkt := enc.Encode(nil)
	if err != nil {
		t.Fatal(err)
	}
	writePkt(pkt, out)

	out.Close()
	enc.Deinit()
}
