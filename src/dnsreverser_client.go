package main

// This file provides the client for DNS handshake + TCP reverser
// without encryption or authentication

import (
    "net"

    "dnshandshake"
    "reverser"
)

func main() {
    // TODO: args should come from command line
    dnsServerAddr := "127.0.0.1:10053"
    dnsQueryDomain := "google.com"

    clientNet := "tcp"
    clientListenAddr := ":10000"

    revNet := "tcp"
    revListenAddr := ":10010"

    // my address for connection
    myIP := "127.0.0.1"
    var myPort uint16 = 10010

    // ==================================
    // TODO: handle encoding error
    query, _ := dnshandshake.EncodeAddr(net.ParseIP(myIP), myPort)

    // TODO: handle potential error
    go func() {
        dnshandshake.SendRequest(query, dnsQueryDomain, dnsServerAddr)
    } ()

    // TODO: handle error
    reverser.StartClient(
        clientNet, clientListenAddr,
        revNet, revListenAddr,
    )
}
