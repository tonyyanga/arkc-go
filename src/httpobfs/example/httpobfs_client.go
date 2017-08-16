package main

import (
    "httpobfs"
    "log"
)

func main() {
    err := httpobfs.StartHTTPObfsClient(
               "http://127.0.0.1:20010",
               "127.0.0.1:20000",
           )
    if err != nil {
        log.Printf("Error occurred: %v\n", err)
    }
}
