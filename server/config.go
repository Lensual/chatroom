package main

import (
	"encoding/json"
	"io/ioutil"

	"github.com/Lensual/chatroom/server/channel"
)

type Config struct {
	Listen   string           `json:"listen"`
	Channels []channel.Config `json:"channels"`
}

var config = Config{}

func loadConfig(filename string) error {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}

	err = json.Unmarshal(data, &config)
	if err != nil {
		return err
	}
	return nil
}
