package client

//CurrentProtocolVersion 当前协议版本 魔数
const CurrentProtocolVersion byte = 0x01

//Header 协议头部
type Header struct {
	ProtocolVersion byte
	Command         CommandType
	Length          uint16
}

//CommandType 指令类型
type CommandType byte

//指令类型
const (
	CmdHello         CommandType = 0x00
	CmdCreateChannel CommandType = 0x01
)
