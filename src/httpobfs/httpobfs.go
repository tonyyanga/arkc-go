// Package httpobfs is similar to meek by Tor project. It can convert channel of data
// into HTTP connection stream
package httpobfs

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

