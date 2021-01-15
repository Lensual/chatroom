# chatroom

自己的聊天室，主要配合自制的网络对讲机实现频道通联

支持Windows（mingw64）和Linux环境

# Build

## Clone

```
git clone https://github.com/Lensual/chatroom
cd chatroom 
git submodule update --init
```

## Install Requirement

Debian/Ubuntu

```
sudo apt install -y git make gcc autoconf yasm pkg-config libtool
```

MSYS2(MinGW64)

```
//TODO
```

## Build Library

```
cd lib
./build_opus.sh
./build_ffmpeg.sh
```

## Build Project

```
go build ./server
```