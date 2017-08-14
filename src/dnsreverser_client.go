package main

// This file provides the client for DNS handshake + TCP reverser
// without encryption or authentication

import (
    "net"
    "log"
    "time"
    "strconv"

    "dnshandshake"
    "reverser"
)

func main() {
    // TODO: args should come from command line
    dnsServerAddr := "127.0.0.1:10053"
    dnsQueryDomain := "baidu.com"

    clientNet := "tcp"
    clientListenAddr := ":10000"

    revNet := "tcp"
    revListenAddr := ":10010"

    // ==================================
    // my address for connection
    myIP, err := dnshandshake.GetExternalIP()
    if err != nil {
        log.Fatalf("Error occurred when getting external IP address: %v\n", err)
    }

    _, p, err := net.SplitHostPort(revListenAddr)
    if err != nil {
        log.Fatalf("Error occurred when parsing revListenAddr: %v\n", err)
    }

    port, err := strconv.ParseUint(p, 10, 16)
    if err != nil {
        log.Fatalf("Error occurred when parsing revListenAddr: %v\n", err)
    }

    var myPort uint16 = uint16(port) 

    // TODO: handle encoding error
    query, _ := dnshandshake.EncodeAddr(myIP, myPort)

    for {
        // TODO: handle potential error
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
