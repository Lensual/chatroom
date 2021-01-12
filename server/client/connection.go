package client

import (
	"encoding/binary"
	"net"
)

//Connection 客户端连接
type Connection struct {
	tcpConn net.Conn
}

//Hello 发送Hello
func (conn *Connection) Hello() error {
	header := Header{
		ProtocolVersion: CurrentProtocolVersion,
		Command:         CmdHello,
		Length:          0,
	}
	err := binary.Write(conn.tcpConn, binary.BigEndian, header)
	if err != nil {
		conn.onErrHandle()
	}
	return err
}

//CreateChannel 创建频道
func (conn *Connection) CreateChannel() error {
	header := Header{
		ProtocolVersion: CurrentProtocolVersion,
		Command:         CmdCreateChannel,
		Length:          0, //TODO,
	}
	err := binary.Write(conn.tcpConn, binary.BigEndian, header)
	if err != nil {
		conn.onErrHandle()
	}
	return err
}

//onErrHandle 错误处理
func (conn *Connection) onErrHandle() {
	_ = conn.tcpConn.Close()
}

//NewConnection 新客户端连接
func NewConnection(tcpConn net.Conn) (*Connection, error) {
	conn := Connection{
		tcpConn: tcpConn,
	}
	err := conn.Hello()

	return &conn, err
}
