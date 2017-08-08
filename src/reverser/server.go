// Package reverser allows reversal of TCP connection direction
package reverser

import (
    //"net"
)

// This function starts the server of the reverser and will block unless an error occurred
// Run this fucnction in a goroutine
func StartServer(
    targetNet string, targetAddr string, // connect address to dump input from reverser client
    revNet string, revConnAddr string, // address to connect to reverser client
) error {

    return nil
}
