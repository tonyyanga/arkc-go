package httpobfs

import (
    "math/rand"
)

// This function generates a random session id
func GenerateRandSessionID() []byte {
    buf := make([]byte, SessionIDLength)
    rand.Read(buf)
    return buf
}

