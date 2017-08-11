export GOPATH=${CURDIR}

GOBUILDFLAGS =

BIN_DIR = bin

TEST_BIN_DIR = ${BIN_DIR}/test

all: test

test: reverser_test dnshandshake_test

reverser_test: ${TEST_BIN_DIR}/reverserClient ${TEST_BIN_DIR}/reverserServer

dnshandshake_test: ${TEST_BIN_DIR}/dns_server

${TEST_BIN_DIR}/dns_server: src/dnshandshake/*.go src/dnshandshake/test/dns_server.go
	go get github.com/miekg/dns
	go build ${GOBUILDFLAGS} -o ${TEST_BIN_DIR}/dns_server src/dnshandshake/test/dns_server.go

${TEST_BIN_DIR}/reverserClient: src/reverser/*.go src/reverser/test/reverserClient.go
	go build ${GOBUILDFLAGS} -o ${TEST_BIN_DIR}/reverserClient src/reverser/test/reverserClient.go

${TEST_BIN_DIR}/reverserServer: src/reverser/*.go src/reverser/test/reverserServer.go
	go build ${GOBUILDFLAGS} -o ${TEST_BIN_DIR}/reverserServer src/reverser/test/reverserServer.go

clean:
	rm -rf ${BIN_DIR}/*
	rm -rf pkg/*
	rm -rf github.com/

.PHONY: all test reverser_test dnshandshake_test clean
