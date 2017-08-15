package httpobfs

import (
    //"net"
)

// This file provides a simple adapter for httpobfs server that issues TCP connections
// to targetAddr
func startObfsServerImpl(targetAddr string) (*HTTPServer, error) {
    return nil, nil
}

func StartHTTPObfsServer(listenAddr string, targetAddr string) error {
    server, err := startObfsServerImpl(targetAddr)
    if err != nil {
        return err
    }

    return server.ListenAndServe(listenAddr)
}

func StartHTTPSObfsServer(listenAddr, certPath, keyPath string, targetAddr string) error {
    server, err := startObfsServerImpl(targetAddr)
    if err != nil {
        return err
    }

    return server.ListenAndServeTLS(listenAddr, certPath, keyPath)
}


