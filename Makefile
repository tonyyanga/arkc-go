export GOPATH=${CURDIR}

GOBUILDFLAGS =

BIN_DIR = bin

all: reverser

reverser: bin/reverserClient bin/reverserServer

${BIN_DIR}/reverserClient: src/reverser/*.go src/reverserClient.go
	go build ${GOBUILDFLAGS} -o ${BIN_DIR}/reverserClient src/reverserClient.go

${BIN_DIR}/reverserServer: src/reverser/*.go src/reverserServer.go
	go build ${GOBUILDFLAGS} -o ${BIN_DIR}/reverserServer src/reverserServer.go

clean:
	rm -f ${BIN_DIR}/*

.PHONY: all reverser clean
