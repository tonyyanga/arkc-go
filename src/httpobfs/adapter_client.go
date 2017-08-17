package httpobfs

import (
    "net"
    "log"
)

// This file provides a simple httpobfs client that listens at listenAddr and forward
// the requests to url provided.

// handle incoming connection and notify httpobfs module
// Spec: 1. Issues NEW_SESSION when connection starts
// Spec: 2. Calls UnregisterID when connection closes
func handleConn(client *HTTPClient, conn net.Conn) {
    // Generate session id for this conn
    var id string
    for {
        buf := GenerateRandSessionID()
        client.mux.RLock()
        _, exists := client.ChanMap[string(buf)]
        client.mux.RUnlock()
        if !exists {
            id = string(buf)
            break
        }
    }

    // Register ID with HTTP client
    pair := client.RegisterID(id)
    sendChan := pair.SendChan
    recvChan := pair.RecvChan

    // Issue NEW_SESSION to register with the other side
    block := &DataBlock{
        SessionID: []byte(id),
        Length: NEW_SESSION,
    }
    sendChan <- block

    // errChan is used to handle cases when connection ends
    // Spec: err == nil means END_SESSION detected; otherwise, an error has
    // occurred in the incoming connection
    errChan := make(chan error, 1)

    // Start regular data writer
    go handleConnRead([]byte(id), conn, sendChan, errChan)

    // Start regular data reader
    go handleConnWrite([]byte(id), conn, recvChan, errChan)

    // Error received, close conn
    err := <-errChan
    conn.Close()
    if err != nil {
        // Error happens on local side, issues END_SESSION
        block := &DataBlock{
            SessionID: []byte(id),
            Length: END_SESSION,
        }
        sendChan <- block
    }

    client.UnregisterID(id)
}

// Start a simple HTTPObfs client
func StartHTTPObfsClient(url string, listenAddr string) error {
    listener, err := net.Listen("tcp", listenAddr)
    if err != nil {
        log.Printf("Error occurred when listening for HTTP obfs reqeusts: %v\n", err)
        return err
    }

    defer listener.Close()

    client := &HTTPClient{}

    // Prepare HTTP obfs client
    client.Start(url)

    for {
        conn, err := listener.Accept()
        if err != nil {
            log.Printf("Error occurred when accepting HTTP obfs requests: %v\n", err)
        }

        go handleConn(client, conn)
    }
}
