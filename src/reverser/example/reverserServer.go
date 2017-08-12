package main

import (
    "reverser"
)

func main() {
    reverser.StartServer("tcp", "www.google.com:80", "tcp", "127.0.0.1:11000")
}
