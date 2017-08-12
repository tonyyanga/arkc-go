export GOPATH=${CURDIR}

GOBUILDFLAGS =

BIN_DIR = bin

TEST_BIN_DIR = ${BIN_DIR}/test

all: prep dnsreverser test

prep:
	go get github.com/miekg/dns

dnsreverser: prep ${BIN_DIR}/dnsreverser_client

${BIN_DIR}/dnsreverser_client: src/dnsreverser_client.go src/dnshandshake/*.go src/reverser/*.go
	go build ${GOBUILDFLAGS} -o ${BIN_DIR}/dnsreverser_client src/dnsreverser_client.go

test: prep reverser_test dnshandshake_test

reverser_test: ${TEST_BIN_DIR}/reverserClient ${TEST_BIN_DIR}/reverserServer

dnshandshake_test: ${TEST_BIN_DIR}/dns_server ${TEST_BIN_DIR}/test_dns_client

${TEST_BIN_DIR}/dns_server: src/dnshandshake/*.go src/dnshandshake/example/dns_server.go
	go build ${GOBUILDFLAGS} -o ${TEST_BIN_DIR}/dns_server src/dnshandshake/example/dns_server.go

${TEST_BIN_DIR}/test_dns_client: src/dnshandshake/*.go src/dnshandshake/example/test_dns_client.go
	go build ${GOBUILDFLAGS} -o ${TEST_BIN_DIR}/test_dns_client src/dnshandshake/example/test_dns_client.go

${TEST_BIN_DIR}/reverserClient: src/reverser/*.go src/reverser/example/reverserClient.go
	go build ${GOBUILDFLAGS} -o ${TEST_BIN_DIR}/reverserClient src/reverser/example/reverserClient.go

${TEST_BIN_DIR}/reverserServer: src/reverser/*.go src/reverser/example/reverserServer.go
	go build ${GOBUILDFLAGS} -o ${TEST_BIN_DIR}/reverserServer src/reverser/example/reverserServer.go

clean:
	rm -rf ${BIN_DIR}/*
	rm -rf pkg/*
	rm -rf src/github.com/

.PHONY: all test reverser_test dnshandshake_test clean
