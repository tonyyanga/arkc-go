package httpobfs

import (
    "net"
    //"log"
)

// This file provides a simple adapter for httpobfs server that issues TCP connections
// to targetAddr

type obfsServerContext struct {
    server *HTTPServer
    targetAddr string
}

// handle incoming blocks, end with END_SESSION
func (ctx *obfsServerContext) handleServerSession(id string) {
    ctx.server.mux.RLock()
    pair, exists := ctx.server.ChanMap[id]
    ctx.server.mux.RUnlock()
    if !exists {
        panic ("cannot find channel pair on map")
    }
    sendChan := pair.SendChan
    recvChan := pair.RecvChan

    // Create connection to target
    conn, err := net.Dial("tcp", ctx.targetAddr)
    if err != nil {
        // connection fails, send END_SESSION
        block := &DataBlock{
            SessionID: []byte(id),
            Length: END_SESSION,
        }
        sendChan <- block
        return
    }

    errChan := make(chan error, 1)

    // Start read goroutine
    go handleConnWrite([]byte(id), conn, recvChan, errChan)

    // Start write goroutine
    go handleConnRead([]byte(id), conn, sendChan, errChan)

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

func startObfsServerImpl(targetAddr string) *HTTPServer {
    server := &HTTPServer{}

    ctx := &obfsServerContext{
        server: server,
        targetAddr: targetAddr,
    }

    server.connHandler = func(id string) {
        ctx.handleServerSession(id)
    }

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


