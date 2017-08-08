// Package reverser allows reversal of TCP connection direction
// It should create a control connection and then maintain connections that map 
// one to one with connections from the client
package reverser

import (
    "net"
)

// Denotes the state of reverser client, i.e. how many connections from client
type state struct {
    clientCount int
}

func handleCtrlConn(conn net.Conn) {

}

// This function starts the client of the reverser and will block unless an error occurred
// Run this fucnction in a goroutine
func StartClient(
    clientNet string, clientListenAddr string, // listen address, for client to connect
    revNet string, revListenAddr string, // listen address for reverser server to connect 
) error {
    // Listen on both interfaces for incoming connections
    clientListener, errC := net.Listen(clientNet, clientListenAddr)
    if errC != nil {
        // TODO: proper error logging
        return errC
    }

    revListener, errR := net.Listen(revNet, revListenAddr)
    if errR != nil {
        // TODO: proper error logging
        return errR
    }

    cliCh := make(chan net.Conn)
    revCh := make(chan net.Conn)

    // ctrlCh is a channel used separately for control connections
    ctrlCh := make(chan net.Conn)

    pairMatcher := connPairMatcher{
        clientConnChan: make(chan net.Conn, 1),
        reverserConnChan: make(chan net.Conn, 1),
        connPairChan: make(chan connPair),
    }

    // Accept for reverser server side connections
    // The first reverser side connection is always the control
    // connection, used to start other data connections
    go func() {
        connCount := true
        for {
            revConn, err := revListener.Accept()
            if err != nil {
                // TODO: error handling
            }

            // TODO: logging
            if connCount {
                ctrlCh <- revConn
                connCount = false;
            } else {
                revCh <- revConn
            }
        }
    } ()

    // Accept for client side connections
    go func() {
        for {
            cliConn, err := clientListener.Accept()
            if err != nil {
                // TODO: error handling
            }

            // TODO: logging
            cliCh <- cliConn
        }
    } ()

    // Connections are handled by separated functions
    for {
        select {
        case conn := <-cliCh:
            go handleClientConn(conn, &pairMatcher)
        case conn := <-revCh:
            go handleRevConn(conn, &pairMatcher)
        case conn := <-ctrlCh:
            go handleCtrlConn(conn)
        }
    }

    return nil
}
