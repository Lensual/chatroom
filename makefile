GO_VERSION = $(shell go version)
BUILD_TIME = $(shell date "+%F %T")
COMMIT_HASH = $(shell git show -s --format="%H %s")

# LDFLAGS= -ldflags "-X server.commitHash=${COMMIT_HASH}	-X server.buildTime=${BUILD_TIME} -X server.goVersion=${GO_VERSION}'" \

all: chatroom-server

chatroom-server:
	go build -ldflags '-X "main.commitHash=${COMMIT_HASH}" -X "main.buildTime=${BUILD_TIME}" -X "main.goVersion=${GO_VERSION}"' \
	-o chatroom-server.exe ./server