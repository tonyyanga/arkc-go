package main

import (
    "dnshandshake"
    "fmt"
)

func main() {
    result, err := dnshandshake.SendRequest("www", "google.com", "8.8.8.8:53")
    if err != nil {
        fmt.Printf("ERROR! %v\n", err)
    } else {
        fmt.Printf("Result: %v\n", result)
    }
}

