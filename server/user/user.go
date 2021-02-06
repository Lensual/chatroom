package user

import (
	"net"

	"github.com/Lensual/chatroom/codec"
)

type User struct {
	Name       string
	Connection net.Conn
	decoder    codec.Decoder
}

func (this *User) Init() {

}
