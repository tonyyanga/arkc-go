package main

import (
    "httpobfs"
    "log"
)

func main() {
    err := httpobfs.StartHTTPObfsServer(
               "127.0.0.1:20010",
               "127.0.0.1:3128",
           )
    if err != nil {
        log.Printf("Error occurred: %v\n", err)
    }
}
