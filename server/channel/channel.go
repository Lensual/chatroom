package channel

import (
	. "github.com/Lensual/chatroom/server/user"
)

var channels []Channel
var channel_count byte

type Channel struct {
	Name  string
	Users []User
	// Encoder codec.Codec
}

//初始化频道模块
func InitChannel() {
	channel_count = 0
	channels = make([]Channel, 256)

}

//新增频道
func NewChannel(name string) {
	channels[int(channel_count)] = Channel{
		Name:  name,
		Users: make([]User, 0),
	}
}

//清理频道
func ClearChannel(index byte) {
	for _, v := range channels[index].Users {
		v.Connection.Close()
	}
	channels[index] = Channel{}
}
