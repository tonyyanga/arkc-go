package main

import (
    "dnshandshake"
    "fmt"
)

func main() {
    ip, err := dnshandshake.GetExternalIP()
    if err != nil {
        fmt.Printf("error when getting external ip, %v\n", err)
        return
    }
    fmt.Printf("My IP address is %v\n", ip)
    result, err := dnshandshake.SendRequest("www", "google.com", "8.8.8.8:53")
    if err != nil {
        fmt.Printf("ERROR! %v\n", err)
    } else {
        fmt.Printf("Result: %v\n", result)
    }
}

