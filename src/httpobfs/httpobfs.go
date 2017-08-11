// Package httpobfs is similar to meek by Tor project. It can convert channel of data
// into HTTP connection stream
package httpobfs

import (
    "io"
    "errors"
    "encoding/binary"
)

// This file provides common data structures used across this package

const SessionIDLength = 16 // Length of session id

type SessionState int

// Enum of possible session state
const (
    NEW_SESSION int = -1
    END_SESSION int = -2

    SESSION_STATE_NUM = iota
)

// DataBlock is used to communicate between different interfaces
type DataBlock struct {
    SessionID []byte // Every conceptual stream has its unique SessionID
    Length int // If positive, length of Data; if negative, follow above enum
    Data []byte
}

func constructDataBlock(body io.Reader) (*DataBlock, error) {
    block := &DataBlock{
        SessionID: make([]byte, SessionIDLength),
    }

    // First read Session ID from payload
    n, err := body.Read(block.SessionID)
    if n < SessionIDLength || err != nil {
        return nil, errors.New("Error when reading session id")
    }

    // Next read Data length
    err = binary.Read(body, binary.BigEndian, block.Length)
    if err != nil {
        return nil, errors.New("Error when reading payload length")
    }

    if block.Length > 0 {
        // Finally read Data
        block.Data = make([]byte, block.Length)
        n, err = body.Read(block.Data)
        if n < block.Length || err != nil {
            return nil, errors.New("Error when reading data")
        }
    }

    return block, nil
}

