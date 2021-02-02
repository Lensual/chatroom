package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"time"

	"github.com/Lensual/chatroom/codec"
)

func main() {
	start := time.Now() // 获取当前时间
	encode()
	elapsed := time.Since(start)
	fmt.Println("编码耗时：", elapsed)
}

func encode() {
	enc := codec.OpusEncoder{}
	err := enc.Init(codec.Mono, 16000, 64000)
	if err != nil {
		panic(err)
	}

	data, err := ioutil.ReadFile("test.pcm")
	if err != nil {
		panic(err)
	}

	out, err := os.Create("test.opus")
	if err != nil {
		panic(err)
	}

	step := enc.GetFrameSize() //帧大小为步进
	for i := 0; i < len(data); i += step {
		if i+step > len(data) {
			break
		}
		frame := data[i : i+step]
		err, pkt := enc.Encode(&frame)
		if err != nil {
			panic(err)
		}

		// fmt.Printf("%v\n", pkt)

		for _, v := range *pkt {
			_, err := out.Write(v)
			if err != nil {
				panic(err)
			}
		}
	}

	out.Close()
}
