package httpobfs

import (
    "net"
    "sync"
    "log"
)

// This file provides a simple httpobfs client that listens at listenAddr and forward
// the requests to url provided.

type obfsClientContext struct {
    client *HTTPClient
    mux sync.RWMutex
}

func (ctx *obfsClientContext) handleConn(conn net.Conn) {
    // Generate session id for this conn
    var id string
    for {
        buf := GenerateRandSessionID()
        ctx.mux.RLock()
        _, exists := ctx.client.ChanMap[string(buf)]
        ctx.mux.RUnlock()
        if !exists {
            id = string(buf)
            break
        }
    }

    // Register ID with HTTP client
    pair := ctx.client.RegisterID(id)
    sendChan := pair.SendChan
    recvChan := pair.RecvChan

    // Issue NEW_SESSION to register with the other side
    block := &DataBlock{
        SessionID: []byte(id),
        Length: NEW_SESSION,
    }
    sendChan <- block

    errChan := make(chan error, 1)

    // Start regular data writer
    go handleConnRead([]byte(id), conn, sendChan, errChan)

    // Start regular data reader
    go handleConnWrite([]byte(id), conn, recvChan, errChan)

    err := <-errChan
    // Error received, close conn
    conn.Close()
    if err != nil {
        // Error happens on local side, need END_SESSION
        block := &DataBlock{
            SessionID: []byte(id),
            Length: END_SESSION,
        }
        sendChan <- block
    }

    ctx.client.UnregisterID(id)
}

func StartHTTPObfsClient(url string, listenAddr string) error {
    listener, err := net.Listen("tcp", listenAddr)
    if err != nil {
        log.Printf("Error occurred when listening for HTTP obfs reqeusts: %v\n", err)
        return err
    }

    defer listener.Close()

    client := &HTTPClient{}

    ctx := &obfsClientContext{
        client: client,
    }

    // Prepare HTTP obfs client
    client.Start(url)

    for {
        conn, err := listener.Accept()
        if err != nil {
            log.Printf("Error occurred when accepting HTTP obfs requests: %v\n", err)
        }

        go ctx.handleConn(conn)
    }
}
