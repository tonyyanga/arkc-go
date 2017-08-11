// Package reverser allows reversal of TCP connection direction
package reverser

import (
    "net"
    "bufio"
    "io"
    "log"
)

//===========================================================================
// revServer contains properties of the reverser server and does actual
// networking operations
type revServer struct {
    // Default methods to describe connections
    targetNet string
    targetAddr string
    revNet string
    revConnAddr string

    // These two functions can be used to read / write to customized io Channel
    dialRevAddr func() (io.ReadWriteCloser, error)
    dialTargetAddr func() (io.ReadWriteCloser, error)

    state *connState
}

// Default function to get connection to revConnAddr
func (s *revServer) getDefaultDialRevAddr() func() (io.ReadWriteCloser, error) {
    return func() (io.ReadWriteCloser, error) {
        return net.Dial(s.revNet, s.revConnAddr)
    }
}

// Default function to get connection to targetAddr
func (s *revServer) getDefaultDialTargetAddr() func() (io.ReadWriteCloser, error) {
    return func() (io.ReadWriteCloser, error) {
        return net.Dial(s.targetNet, s.targetAddr)
    }
}

// This function does an individual connection and send 1 to finishChan
// channel
func (s *revServer) newConnection(finishChan chan byte) {
    revConn, err := s.dialRevAddr()
    if err != nil {
        log.Printf("Error occurred when connecting to reverser client: %v\n", err)
        return
    }
    defer revConn.Close()
    defer func() {finishChan <- 1} ()

    targetConn, err := s.dialTargetAddr()
    if err != nil {
        log.Printf("Error occurred when connecting to target address: %v\n", err)
        return
    }
    defer targetConn.Close()


    errChan := make(chan error)

    go connCopy(revConn, targetConn, errChan)
    go connCopy(targetConn, revConn, errChan)

    err = <-errChan
    if err != nil {
        log.Printf("Error occurred when copying data: %v\n", err)
        return
    }
    log.Println("Close a connection pair")
}

// This function starts individual goroutines which do connections
// according to state updates
func (s *revServer) connect() {
    finishChan := make(chan byte)
    var activeConn byte = 0
    var requiredNum byte = 0

    for {
        // Receive update about connection number
        select{
        case requiredNum = <-s.state.GetUpdateChan():

        case <-finishChan:
            activeConn--
        }

        if requiredNum <= activeConn {
            continue
        } else {
            var i byte
            for i = 0; i < requiredNum - activeConn; i++ {
                go s.newConnection(finishChan)
            }
            activeConn = requiredNum
        }
    }
}

// Start the reverser server
func (s *revServer) Start() error {
    // First start the control connection
    conn, err := s.dialRevAddr()
    if err != nil {
        log.Printf("Error occurred when creating control connection: %v\n", err)
        return err
    }
    defer conn.Close()

    // Connect to reverser client based on client's need
    go s.connect()

    for {
        r := bufio.NewReader(conn)
        requiredNum, err := r.ReadByte()
        if err != nil {
            log.Printf("Error occurred when reading from control connection: %v\n", err)
            return err
        }

        log.Printf("Receive request for %d connections\n", requiredNum)
        s.state.UpdateClientCount(requiredNum)
    }

    return nil
}
//=============================== END =================================

// This function starts the server of the reverser and will block unless an error occurred
// Run this fucnction in a goroutine
func StartServer(
    tNet string, tAddr string, // connect address to dump input from reverser client
    rNet string, rConnAddr string, // address to connect to reverser client
) error {

    server := &revServer{
        targetNet: tNet,
        targetAddr: tAddr,
        revNet: rNet,
        revConnAddr: rConnAddr,

        state: &connState{
            clientCount: 0,
            updateChan: make(chan byte, 5),
        },
    }

    server.dialRevAddr = server.getDefaultDialRevAddr()
    server.dialTargetAddr = server.getDefaultDialTargetAddr()

    return server.Start()
}

// This function starts the server of the reverser and will block unless an error occurred
// Run this fucnction in a goroutine
func StartServerWithDialerFunc(
    targetDialer func() (io.ReadWriteCloser, error), // dialer function for channel to target
    revDialer func() (io.ReadWriteCloser, error), // dialer function for channel to reverser client
) error {

    server := &revServer{
        dialRevAddr: revDialer,
        dialTargetAddr: targetDialer,

        state: &connState{
            clientCount: 0,
            updateChan: make(chan byte, 5),
        },
    }

    return server.Start()
}
