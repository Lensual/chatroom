package main

import (
	"flag"
	"fmt"
	"net"

	"github.com/Lensual/chatroom/codec"
	ClientConn "github.com/Lensual/chatroom/server/client"
)

///build info
const (
	BuildVersion string = "not set"
	BuildTime    string = "not set"
	BuildName    string = "not set"
	CommitID     string = "not set"
)

func main() {
	var showVer bool

	flag.BoolVar(&showVer, "v", false, "show build version")

	flag.Parse()

	if showVer {
		// Printf( "build name:\t%s\nbuild ver:\t%s\nbuild time:\t%s\nCommitID:%s\n", BuildName, BuildVersion, BuildTime, CommitID )
		fmt.Printf("build name:\t%s\n", BuildName)
		fmt.Printf("build ver:\t%s\n", BuildVersion)
		fmt.Printf("build time:\t%s\n", BuildTime)
		fmt.Printf("Commit ID:\t%s\n", CommitID)
		return
	}

	opus := codec.OpusDecoder{}

	_ = opus.Init(1, 22050)

	tcpListener, err := net.Listen("tcp", ":7000")
	if err != nil {
		panic(err)
	}
	for {
		tcpConn, err := tcpListener.Accept()
		if err != nil {
			panic(err)
		}
		conn, err := ClientConn.NewConnection(tcpConn)
		if err == nil {
			go doWork(conn)
		}
	}

}

func doWork(conn *ClientConn.Connection) {

}
