package httpobfs

import (
    "net"
    "log"
)

// This file provides a simple httpobfs client that listens at listenAddr and forward
// the requests to url provided.

type obfsClientContext struct {
    client *HTTPClient
    dispatchMap map[string] chan *DataBlock
}

func (ctx *obfsClientContext) handleConn(conn net.Conn) {
    // Generate session id for this conn
    var id string
    for {
        buf := GenerateRandSessionID()
        _, exists := ctx.dispatchMap[string(buf)]
        if !exists {
            id = string(buf)
            break
        }
    }

    // Add myself to dispatchMap
    recvChan := make(chan *DataBlock, 10)
    ctx.dispatchMap[id] = recvChan

    // Issue NEW_SESSION to register with the other side
    block := &DataBlock{
        SessionID: []byte(id),
        Length: NEW_SESSION,
    }
    ctx.client.SendChan <- block

    errChan := make(chan error, 1)

    // Start regular data writer
    go handleConnRead([]byte(id), conn, ctx.client.SendChan, errChan)

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
        ctx.client.SendChan <- block
        delete(ctx.dispatchMap, id)
    }
}

func (ctx *obfsClientContext) dispatch() {
    for {
        block := <-ctx.client.RecvChan
        sessionID := string(block.SessionID)

        if block.Length == NEW_SESSION {
            // Illegal input
            log.Printf("Error: illegal input from server\n")
            continue
        }

        chanToSend, exists := ctx.dispatchMap[sessionID]
        if !exists {
            log.Printf("Error: cannot find channel in dispatch map, packet dropped\n")
            continue
        }
        chanToSend <- block

        if block.Length == END_SESSION {
            // remove channel from dispatchMap
            delete(ctx.dispatchMap, sessionID)
        }
    }
}

func StartHTTPObfsClient(url string, listenAddr string) error {
    listener, err := net.Listen("tcp", listenAddr)
    if err != nil {
        log.Printf("Error occurred when listening for HTTP obfs reqeusts: %v\n", err)
        return err
    }

    defer listener.Close()

    client := &HTTPClient{
        RecvChan: make(chan *DataBlock, 50),
        SendChan: make(chan *DataBlock, 50),
    }

    ctx := &obfsClientContext{
        dispatchMap: make(map[string] chan *DataBlock),
        client: client,
    }

    // Start HTTP obfs client
    go client.Start(url)

    // Start connection dispatcher
    go ctx.dispatch()

    for {
        conn, err := listener.Accept()
        if err != nil {
            log.Printf("Error occurred when accepting HTTP obfs requests: %v\n", err)
        }

        go ctx.handleConn(conn)
    }

}
