package httpobfs

import (
    "net"
    //"log"
)

// This file provides a simple adapter for httpobfs server that issues TCP connections
// to targetAddr

// Stores Server side context
type obfsServerContext struct {
    server *HTTPServer
    targetAddr string
}

// Initiate connection to targetAddr and handle blocks
// Specs: 1. Should receive no new session from sendChan
//        2. End with END_SESSION and close connections
func (ctx *obfsServerContext) handleServerSession(id string) {
    // Get channel pair
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

    // errChan is used to handle connection closing or END_SESSION
    // Spec: nil means END_SESSION detected; otherwise, error has happened on
    // connection to targetAddr
    errChan := make(chan error, 1)

    // Start read goroutine
    go handleConnWrite([]byte(id), conn, recvChan, errChan)

    // Start write goroutine
    go handleConnRead([]byte(id), conn, sendChan, errChan)

    // If err received from error channel, close connection
    err = <-errChan
    conn.Close()

    if err != nil {
        // Error happens with targetAddr, issue END_SESSION
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

    server.ConnHandler = func(id string) {
        ctx.handleServerSession(id)
    }

    return server
}

// Start a simple HTTP obfs server without SSL
func StartHTTPObfsServer(listenAddr string, targetAddr string) error {
    server := startObfsServerImpl(targetAddr)
    return server.ListenAndServe(listenAddr)
}

// Start a simple HTTPS obfs server, TLS used
func StartHTTPSObfsServer(listenAddr, certPath, keyPath string, targetAddr string) error {
    server := startObfsServerImpl(targetAddr)
    return server.ListenAndServeTLS(listenAddr, certPath, keyPath)
}


