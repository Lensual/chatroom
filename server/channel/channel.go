package channel

import (
	"github.com/Lensual/chatroom/codec"
	. "github.com/Lensual/chatroom/server/user"
)

var channels map[string]Channel

type Channel struct {
	Name    string
	Users   []User
	Encoder codec.Encoder
}

//初始化频道模块
func InitChannel(config []Config) {
	channels = make(map[string]Channel, len(config))
	for _, v := range config {
		CreateChannel(v)
	}
}

//新增频道
func CreateChannel(config Config) {
	channels[config.Name] = Channel{
		Name:  config.Name,
		Users: make([]User, 0),
	}
}

//清理频道
func ClearChannel(channel *Channel) {
	//TODO
}
