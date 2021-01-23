package main

import (
	"flag"
	"fmt"
	"net"

	"github.com/Lensual/chatroom/server/channel"
	ClientConn "github.com/Lensual/chatroom/server/client"
)

///build info
var (
	commitHash string = "unknown"
	buildTime  string = "unknown"
	goVersion  string = "unknown"
)

func main() {
	var showVer bool
	var configFile string

	flag.BoolVar(&showVer, "v", false, "show build version")
	flag.StringVar(&configFile, "c", "config.js", "config file path")

	flag.Parse()

	if showVer {
		fmt.Printf("Git Commit Hash: %s \n", commitHash)
		fmt.Printf("Build TimeStamp: %s \n", buildTime)
		fmt.Printf("GoLang Version: %s \n", goVersion)
		return
	}

	loadConfig(configFile)

	//初始化频道
	channel.InitChannel(config.Channels)

	//监听
	tcpListener, err := net.Listen("tcp", config.Listen)
	if err != nil {
		panic(err)
	}

	//开始Accept循环
	for {
		tcpConn, err := tcpListener.Accept()
		if err != nil {
			tcpConn.Close()
			continue
		}
		conn, err := ClientConn.NewConnection(tcpConn)
		if err == nil {
			go doWork(conn)
		}
	}

}

func doWork(conn *ClientConn.Connection) {

}
