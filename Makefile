export GOPATH=${CURDIR}

GOBUILDFLAGS =

BIN_DIR = bin

all: reverserClient reverserServer

reverserClient: src/reverser/*.go src/reverserClient.go
	go build ${GOBUILDFLAGS} -o ${BIN_DIR}/reverserClient src/reverserClient.go

reverserServer: src/reverser/*.go src/reverserServer.go
	go build ${GOBUILDFLAGS} -o ${BIN_DIR}/reverserServer src/reverserServer.go
