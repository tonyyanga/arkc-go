package main

import (
    "fmt"

    "dnshandshake"
)

func main() {
    recv := make(chan string, 100)

    go dnshandshake.Serve("google.com", "udp", "127.0.0.1:10053", recv)

    for {
        fmt.Printf("Receive request %v\n", <-recv)
    }
}

