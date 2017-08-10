export GOPATH=${CURDIR}

GOBUILDFLAGS =

BIN_DIR = bin

TEST_BIN_DIR = ${BIN_DIR}/test

all: test

test: reverser_test

reverser_test: ${TEST_BIN_DIR}/reverserClient ${TEST_BIN_DIR}/reverserServer

${TEST_BIN_DIR}/reverserClient: src/reverser/*.go src/reverser/test/reverserClient.go
	go build ${GOBUILDFLAGS} -o ${TEST_BIN_DIR}/reverserClient src/reverser/test/reverserClient.go

${TEST_BIN_DIR}/reverserServer: src/reverser/*.go src/reverser/test/reverserServer.go
	go build ${GOBUILDFLAGS} -o ${TEST_BIN_DIR}/reverserServer src/reverser/test/reverserServer.go

clean:
	rm -rf ${BIN_DIR}/*

.PHONY: all test reverser_test clean
