package main

import (
    "time"

    "reverser"
)

func main() {
    reverser.StartClient("tcp", ":10000", "tcp", "127.0.0.1:11000", 1 * time.Second)
}
