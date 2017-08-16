export GOPATH=${CURDIR}

GOBUILDFLAGS +=

BIN_DIR = bin

TEST_BIN_DIR = ${BIN_DIR}/test

EXTRA_ENV_VAR +=

all: prep dnsreverser test

dynamic: GOBUILDFLAGS += -linkshared
dynamic: dynamic_prep dnsreverser test

dynamic_prep:
	${EXTRA_ENV_VAR} go get github.com/miekg/dns
	go install -buildmode=shared ${GOBUILDFLAGS} std
	go install -buildmode=shared ${GOBUILDFLAGS} github.com/miekg/dns

linux_amd64: EXTRA_ENV_VAR += GOOS=linux GOARCH=amd64
linux_amd64: all

linux_386: EXTRA_ENV_VAR += GOOS=linux GOARCH=386
linux_386: all

prep:
	${EXTRA_ENV_VAR} go get github.com/miekg/dns

dnsreverser: prep ${BIN_DIR}/dnsreverser_client ${BIN_DIR}/dnsreverser_server

${BIN_DIR}/dnsreverser_client: src/dnsreverser_client.go src/dnshandshake/*.go src/reverser/*.go
	${EXTRA_ENV_VAR} go build ${GOBUILDFLAGS} -o ${BIN_DIR}/dnsreverser_client src/dnsreverser_client.go

${BIN_DIR}/dnsreverser_server: src/dnsreverser_server.go src/dnshandshake/*.go src/reverser/*.go
	${EXTRA_ENV_VAR} go build ${GOBUILDFLAGS} -o ${BIN_DIR}/dnsreverser_server src/dnsreverser_server.go

test: prep reverser_test dnshandshake_test httpobfs_test

reverser_test: ${TEST_BIN_DIR}/reverserClient ${TEST_BIN_DIR}/reverserServer

dnshandshake_test: ${TEST_BIN_DIR}/dns_server ${TEST_BIN_DIR}/test_dns_client

httpobfs_test: ${TEST_BIN_DIR}/httpobfs_client ${TEST_BIN_DIR}/httpobfs_server

${TEST_BIN_DIR}/httpobfs_client: src/httpobfs/*.go src/httpobfs/example/httpobfs_client.go
	${EXTRA_ENV_VAR} go build ${GOBUILDFLAGS} -o ${TEST_BIN_DIR}/httpobfs_client src/httpobfs/example/httpobfs_client.go

${TEST_BIN_DIR}/httpobfs_server: src/httpobfs/*.go src/httpobfs/example/httpobfs_server.go
	${EXTRA_ENV_VAR} go build ${GOBUILDFLAGS} -o ${TEST_BIN_DIR}/httpobfs_server src/httpobfs/example/httpobfs_server.go

${TEST_BIN_DIR}/dns_server: src/dnshandshake/*.go src/dnshandshake/example/dns_server.go
	${EXTRA_ENV_VAR} go build ${GOBUILDFLAGS} -o ${TEST_BIN_DIR}/dns_server src/dnshandshake/example/dns_server.go

${TEST_BIN_DIR}/test_dns_client: src/dnshandshake/*.go src/dnshandshake/example/test_dns_client.go
	${EXTRA_ENV_VAR} go build ${GOBUILDFLAGS} -o ${TEST_BIN_DIR}/test_dns_client src/dnshandshake/example/test_dns_client.go

${TEST_BIN_DIR}/reverserClient: src/reverser/*.go src/reverser/example/reverserClient.go
	${EXTRA_ENV_VAR} go build ${GOBUILDFLAGS} -o ${TEST_BIN_DIR}/reverserClient src/reverser/example/reverserClient.go

${TEST_BIN_DIR}/reverserServer: src/reverser/*.go src/reverser/example/reverserServer.go
	${EXTRA_ENV_VAR} go build ${GOBUILDFLAGS} -o ${TEST_BIN_DIR}/reverserServer src/reverser/example/reverserServer.go

clean:
	rm -rf ${BIN_DIR}/*
	rm -rf pkg/*
	rm -rf src/github.com/

.PHONY: all dynamic dynamic_prep linux_amd64 linux_386 test prep clean \
	dnsreverser reverser_test dnshandshake_test httpobfs_test
