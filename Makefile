export GOPATH=${CURDIR}

GOBUILDFLAGS =

REVERSER_FILES = \
	src/reverser/client.go \
	src/reverser/client_sessions.go \
	src/reverser/server.go \

BIN_DIR = bin

.PHONY: all clean

all: reverser

reverser: ${REVERSER_FILES} src/reverserClient.go src/reverserClient.go
	go build ${GOBUILDFLAGS} -o ${BIN_DIR}/reverserClient src/reverserClient.go
	go build ${GOBUILDFLAGS} -o ${BIN_DIR}/reverserServer src/reverserServer.go

clean:
	rm -f bin/*
