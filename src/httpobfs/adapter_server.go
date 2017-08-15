package httpobfs

import (
    "net"
    "log"
)

// This file provides a simple adapter for httpobfs server that issues TCP connections
// to targetAddr

type obfsContext struct {
    server *HTTPServer
    dispatchMap map[string] chan *DataBlock
    targetAddr string
}

// handle incoming blocks, end with END_SESSION
func (ctx *obfsContext) handleServerSession(id string, recvChan chan *DataBlock, sendChan chan *DataBlock) {
    // Create connection to target
    conn, err := net.Dial("tcp", ctx.targetAddr)
    if err != nil {
        // connection fails, send END_SESSION
        delete(ctx.dispatchMap, id)
        block := &DataBlock{
            SessionID: []byte(id),
            Length: END_SESSION,
        }
        sendChan <- block
        return
    }

    errChan := make(chan error, 1)

    // Start read goroutine
    go func() {
        for {
            block := <-recvChan
            if block.Length == END_SESSION {
                // nil means connection closed gracefully
                errChan <- nil
                return
            } else {
                // regular connection data
                _, err := conn.Write(block.Data)
                if err != nil {
                    errChan <- err
                }
            }
        }
    } ()

    // Start write goroutine
    go func() {
        for {
            buf := make([]byte, DefaultBlockLength)
            n, err := conn.Read(buf)
            if err != nil {
                errChan <- err
                return
            }
            block := &DataBlock{
                SessionID: []byte(id),
                Length: n,
                Data: buf,
            }
            sendChan <- block
        }
    } ()

    // If err received from error channel, close connection after END_SESSION
    err = <-errChan
    conn.Close()

    if err != nil {
        // Error happens with target
        block := &DataBlock{
            SessionID: []byte(id),
            Length: END_SESSION,
        }
        sendChan <- block
    }

}

// This function dispatch requests from recvChan and start new goroutines to handle
// individual connections
func (ctx *obfsContext) obfsServerDispatch() {
    for {
        block := <-ctx.server.RecvChan

        sessionID := string(block.SessionID)

        if block.Length == NEW_SESSION {
            // New session from the client side, to be handled in a new goroutine
            recvChan := make(chan *DataBlock, 10)
            ctx.dispatchMap[sessionID] = recvChan
            go ctx.handleServerSession(sessionID, recvChan, ctx.server.SendChan)

            // No need to forward NEW_SESSION to target
            continue
        }

        // If not NEW_SESSION, forward block to channel
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

func startObfsServerImpl(targetAddr string) *HTTPServer {
    server := &HTTPServer{
        RecvChan: make(chan *DataBlock, 50),
        SendChan: make(chan *DataBlock, 50),
    }

    ctx := &obfsContext{
        // dispatchMap is used for dispatching RecvChan messages
        dispatchMap: make(map[string] chan *DataBlock),
        server: server,
        targetAddr: targetAddr,
    }

    go ctx.obfsServerDispatch()

    return server
}

func StartHTTPObfsServer(listenAddr string, targetAddr string) error {
    server := startObfsServerImpl(targetAddr)
    return server.ListenAndServe(listenAddr)
}

func StartHTTPSObfsServer(listenAddr, certPath, keyPath string, targetAddr string) error {
    server := startObfsServerImpl(targetAddr)
    return server.ListenAndServeTLS(listenAddr, certPath, keyPath)
}


