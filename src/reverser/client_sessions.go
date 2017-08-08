package reverser

import (
    "net"
)

// Denotes a pair of connections for service 
type connPair struct {
    ClientConn net.Conn
    ReverserConn net.Conn
}

// Get a pair for client connection submitted, blocking
func getClientConnPair(conn net.Conn) connPair {
    return nil
}

// Get a pair for reverser connection submitted, blocking
func getReverserConnPair(conn net.Conn) connPair {
    return nil
}

func handleClientConn(conn net.Conn) {

}

func handleRevConn(conn net.Conn) {

}
