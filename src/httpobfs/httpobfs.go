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

const DefaultBlockLength = 4096

// Enum of possible session state
const (
    NO_DATA int32 = 0
    NEW_SESSION int32 = -1
    END_SESSION int32 = -2

    SESSION_STATE_NUM = iota
)

// DataBlock is used to communicate between different interfaces
type DataBlock struct {
    SessionID []byte // Every conceptual stream has its unique SessionID
    Length int32 // If positive, length of Data; if negative, follow above enum
    Data []byte
}

func constructDataBlock(body io.Reader) (*DataBlock, error) {
    block := &DataBlock{
        SessionID: make([]byte, SessionIDLength),
    }

    // First read Session ID from payload
    _, err := io.ReadFull(body, block.SessionID)
    if err != nil {
        return nil, err
    }

    // Next read Data length
    err = binary.Read(body, binary.BigEndian, &block.Length)
    if err != nil {
        return nil, errors.New("Error when reading payload length")
    }

    if block.Length > 0 {
        // Finally read Data
        block.Data = make([]byte, block.Length)
        _, err = io.ReadFull(body, block.Data)
        if err != nil {
            return nil, err
        }
    //} else if block.Length < 0 {
    //    log.Printf("Length %v at Session ID %v\n", block.Length, block.SessionID)
    }

    return block, nil
}

