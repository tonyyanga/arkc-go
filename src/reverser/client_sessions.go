package reverser

import (
    "net"
    "sync"
    "io"
)

//======================================================================
// Data structure to match incoming client connections with reverser
// connections
//======================================================================

// Denotes a pair of connections for service 
type connPair struct {
    ClientConn net.Conn
    ReverserConn net.Conn
}

type connPairMatcher struct {
    clientConnChan chan net.Conn
    reverserConnChan chan net.Conn
    connPairChan chan connPair

    mux sync.Mutex
}

// Get a pair for client connection submitted, blocking
func (matcher connPairMatcher) GetClientConnPair(conn net.Conn) connPair {
    // TODO: simplify this by using fewer channels
    matcher.mux.Lock()
    select {
    case revConn := <-matcher.reverserConnChan:
        pair := connPair{
            ClientConn: conn,
            ReverserConn: revConn,
        }
        matcher.connPairChan <- pair
        matcher.mux.Unlock()
        return pair
    default:
        matcher.clientConnChan <- conn
        matcher.mux.Unlock()
        return <-matcher.connPairChan
    }
}

// Get a pair for reverser connection submitted, blocking
func (matcher connPairMatcher) GetReverserConnPair(conn net.Conn) connPair {
    // TODO: simplify this by using fewer channels
    matcher.mux.Lock()
    select {
    case cliConn := <-matcher.clientConnChan:
        pair := connPair{
            ClientConn: cliConn,
            ReverserConn: conn,
        }
        matcher.connPairChan <- pair
        matcher.mux.Unlock()
        return pair
    default:
        matcher.reverserConnChan <- conn
        matcher.mux.Unlock()
        return <-matcher.connPairChan
    }
}

//=============================END=====================================

func connCopy(src net.Conn, dst net.Conn, errChan chan error) {
    _, err := io.Copy(dst, src)
    errChan <- err
}

func handleClientConn(conn net.Conn, matcher connPairMatcher) {
    // Get connPair for this connection
    matcher.GetClientConnPair(conn)

    // do nothing because handleRevConn will handle data on both directions
    return
}

func handleRevConn(conn net.Conn, matcher connPairMatcher) {
    // Get connPair for this connection
    pair := matcher.GetReverserConnPair(conn)

    // error when copying from reverser to client, i.e. EOF from reverser
    err1 := make(chan error)
    // error when copying from client to reverser, i.e. EOF from client
    err2 := make(chan error)

    // Do bidirectional copy on separate goroutines
    go connCopy(pair.ReverserConn, pair.ClientConn, err1)
    go connCopy(pair.ClientConn, pair.ReverserConn, err2)

    // Receive input
    select {
    case err := <-err1:
        if err != nil {
            // TODO: process err
        }
    case err := <-err2:
        if err != nil {
            // TODO: process err
        }
    }

    pair.ReverserConn.Close()
    pair.ClientConn.Close()
}
