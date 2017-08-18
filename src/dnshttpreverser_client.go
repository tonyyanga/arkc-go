package main

// This file provides the client for DNS handshake + TCP reverser
// without encryption or authentication

import (
    "net"
    "time"

    "dnshandshake"
    "reverser"
    "httpobfs"
)

func main() {
    // TODO: args should come from command line
    dnsServerAddr := "127.0.0.1:10053"
    dnsQueryDomain := "baidu.com"

    clientNet := "tcp"
    clientListenAddr := ":10000"

    revNet := "tcp"
    revListenAddr := "127.0.0.1:10010"

    httpObfsListenAddr := "127.0.0.1:20010"
    httpObfsTargetAddr := revListenAddr

    // ==================================
    // TODO: handle encoding error
    query, _ := dnshandshake.EncodeAddr(net.ParseIP("127.0.0.1"), 20010)

    for {
        // TODO: handle potential error
        go func() {
            httpobfs.StartHTTPObfsServer(httpObfsListenAddr, httpObfsTargetAddr)
        } ()

        go func() {
            dnshandshake.SendRequest(query, dnsQueryDomain, dnsServerAddr)
        } ()

        // TODO: handle error
        reverser.StartClient(
            clientNet, clientListenAddr,
            revNet, revListenAddr,
            1 * time.Second,
        )
    }
}
