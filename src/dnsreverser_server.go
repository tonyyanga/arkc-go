package main

// This file provides the server for DNS handshake + TCP reverser
// without encryption or authentication

import (
    "net"
    "strconv"
    "log"

    "dnshandshake"
    "reverser"
)

func main() {
    // TODO: args should come from command line
    dnsListenNet := "udp"
    dnsListenAddr := ":10053"
    dnsQueryDomain := "baidu.com"

    targetNet := "tcp"
    targetAddr := "127.0.0.1:3128"

    // ==================================
    dnsRecvChan := make(chan string, 100)
    errChanDNS := make(chan error)
    errChanReverser := make(chan error)

    go func() {
        errChanDNS <- dnshandshake.Serve(
            dnsQueryDomain,
            dnsListenNet, dnsListenAddr,
            dnsRecvChan,
        )
    } ()

    for {
        select {
        case err := <-errChanDNS:
            log.Fatalf("Error when listening for DNS: %v\n", err)
        case err := <-errChanReverser:
            log.Printf("Error when connecting to reverser: %v\n", err)
        case query := <-dnsRecvChan:
            revIP, revPort, err := dnshandshake.DecodeAddr(query)
            if err != nil {
                log.Printf("Error occurred when handling HTTP query: %v\n", err)
                continue
            }

            revAddr := net.JoinHostPort(revIP.String(), strconv.Itoa(int(revPort)))

            go func() {
                errChanReverser <- reverser.StartServer(
                    targetNet, targetAddr,
                    "tcp", revAddr,
                )
            } ()
        }
    }
}
