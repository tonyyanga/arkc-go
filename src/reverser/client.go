// Package reverser allows reversal of TCP connection direction
package reverser

import (
    //"net"
)

// This function starts the client of the reverser and will block unless an error occurred
// Run this fucnction in a goroutine
func StartClient(
    clientNet string, clientListenAddr string, // listen address, for client to connect
    revNet string, revListenAddr string, // listen address for reverser server to connect 
) error {

    return nil
}
