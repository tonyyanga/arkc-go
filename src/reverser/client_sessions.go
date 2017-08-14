package reverser

import (
    "net"
    "sync"
    "errors"
    "log"
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
func (matcher *connPairMatcher) GetClientConnPair(conn net.Conn) (*connPair, error) {
    // TODO: simplify this by using fewer channels
    // TODO: critical section seems to be flawed
    matcher.mux.Lock()
    select {
    case revConn := <-matcher.reverserConnChan:
        pair := connPair{
            ClientConn: conn,
            ReverserConn: revConn,
        }
        matcher.connPairChan <- pair
        matcher.mux.Unlock()
        return &pair, nil
    default:
        matcher.mux.Unlock()
        matcher.clientConnChan <- conn
        pair, more := <-matcher.connPairChan
        if !more {
            return nil, errors.New("Closed channel")
        } else {
            return &pair, nil
        }
    }
}

// Get a pair for reverser connection submitted, blocking
func (matcher *connPairMatcher) GetReverserConnPair(conn net.Conn) (*connPair, error) {
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
        return &pair, nil
    default:
        matcher.mux.Unlock()
        matcher.reverserConnChan <- conn
        pair, more := <-matcher.connPairChan
        if !more {
            return nil, errors.New("Closed channel")
        } else {
            return &pair, nil
        }
    }
}

// Close the matching pair channel
func (matcher *connPairMatcher) Close() {
    matcher.mux.Lock()
    close(matcher.connPairChan)
    matcher.mux.Unlock()
}

//=============================END=====================================

// Handler for client side connections
func handleClientConn(conn net.Conn, matcher *connPairMatcher, state *connState) {
    state.AddClientCount()
    // Get connPair for this connection
    _, err := matcher.GetClientConnPair(conn)
    if err != nil {
        // pair matcher closed, close conn immediately
        conn.Close()
    }

    // do nothing because handleRevConn will handle data on both directions
    return
}

// Handler for reverser server side connections
func handleRevConn(conn net.Conn, matcher *connPairMatcher, state *connState) {
    // Get connPair for this connection
    pair, err := matcher.GetReverserConnPair(conn)
    if err != nil {
        // pair matcher closed, close connection immediately
        conn.Close()
    }

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
            log.Printf("Error occurred when copying from reverser to client: %v\n", err)
        }
    case err := <-err2:
        if err != nil {
            log.Printf("Error occurred when copying from client to reverser: %v\n", err)
        }
    }

    // Simple no connection reuse model
    pair.ReverserConn.Close()
    pair.ClientConn.Close()
    state.ReduceClientCount()

    log.Println("Close a connection pair")
}
