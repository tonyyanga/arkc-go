// Package reverser allows reversal of TCP connection direction
package reverser

import (
    "net"
    "bufio"
)

// revServer contains properties of the reverser server
type revServer struct {
    targetNet string
    targetAddr string
    revNet string
    revConnAddr string

    state *connState
}

func (s *revServer) newConnection(finishChan chan byte) {

}

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

func (s *revServer) Start() error {
    // First start the control connection
    conn, err := net.Dial(s.revNet, s.revConnAddr)
    if err != nil {
        // TODO: handle error
        return err
    }

    // Connect to reverser client based on client's need
    go s.connect()

    for {
        r := bufio.NewReader(conn)
        requiredNum, err := r.ReadByte()
        if err != nil {
            return err
        }

        s.state.UpdateClientCount(requiredNum)
    }

    return nil
}

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

    return server.Start()
}
