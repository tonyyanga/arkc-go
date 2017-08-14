// Package reverser allows reversal of TCP connection direction
// It should create a control connection and then maintain connections that map 
// one to one with connections from the client
package reverser

import (
    "net"
    "bufio"
    "time"
    "errors"
    "log"
)

// Handler for control connection, when connection ends, error goes to errChan
func handleCtrlConn(conn net.Conn, state *connState, errChan chan error) {
    defer conn.Close()
    w := bufio.NewWriter(conn)
    for {
        requiredNum := state.WaitForUpdate()
        err := w.WriteByte(requiredNum)
        if err != nil {
            log.Printf("Error occurred at control connection: %vi\n", err)
            errChan <- err
            return
        }
        err = w.Flush()
        if err != nil {
            log.Printf("Error occurred at control connection: %v\n", err)
            errChan <- err
            return
        }
    }
}

// This function starts the client of the reverser and will block unless an error occurred
// Will return error if the connection is reset or not received within timeout
func StartClient(
    clientNet string, clientListenAddr string, // listen address, for client to connect
    revNet string, revListenAddr string, // listen address for reverser server to connect 
    timeout time.Duration, // timeout for listening for control connection
) error {
    // Listen on both interfaces for incoming connections
    clientListener, err := net.Listen(clientNet, clientListenAddr)
    if err != nil {
        log.Printf("Cannot listen for client connections: %v\n", err)
        return err
    }
    defer clientListener.Close()

    revListener, err := net.Listen(revNet, revListenAddr)
    if err != nil {
        log.Printf("Cannot listen for server-side connections: %v\n", err)
        return err
    }
    defer revListener.Close()

    cliCh := make(chan net.Conn)
    revCh := make(chan net.Conn)

    // ctrlCh is a channel used separately for control connections
    ctrlCh := make(chan net.Conn)

    errChan := make(chan error)

    pairMatcher := &connPairMatcher{
        clientConnChan: make(chan net.Conn, 1),
        reverserConnChan: make(chan net.Conn, 1),
        connPairChan: make(chan connPair),
    }
    defer pairMatcher.Close()

    state := &connState{
        clientCount: 0,
        updateChan: make(chan byte, 5),
    }

    // Accept for reverser server side connections
    // The first reverser side connection is always the control
    // connection, used to start other data connections
    go func() {
        connCount := true
        for {
            revConn, err := revListener.Accept()
            if err != nil {
                log.Printf("Error occurred when accepting server-side connection: %v\n", err)
                return
            }

            if connCount {
                ctrlCh <- revConn
                connCount = false;
                log.Println("Control connection accepted")
            } else {
                revCh <- revConn
                log.Println("Client connection accepted")
            }
        }
    } ()

    // listen for control connection before accepting connections from client
    select {
    case conn := <-ctrlCh:
        // Receive control connection within timeout
        go handleCtrlConn(conn, state, errChan)
    case <- time.After(timeout):
        return errors.New("Didn't receive connection before timeout")
    }

    // Accept for client side connections after control connection is ready
    go func() {
        for {
            cliConn, err := clientListener.Accept()
            if err != nil {
                log.Printf("Error occurred when accepting client connection: %v\n", err)
                return
            }

            cliCh <- cliConn
            log.Println("Reverser connection accepted")
        }
    } ()

    // Connections are handled by separated functions
    for {
        select {
        case conn := <-cliCh:
            go handleClientConn(conn, pairMatcher, state)
        case conn := <-revCh:
            go handleRevConn(conn, pairMatcher, state)
        case err := <-errChan:
            return err
        }
    }

    return nil
}
