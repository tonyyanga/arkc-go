package httpobfs

import (
    "math/rand"
    "net"
)

// This function generates a random session id
func GenerateRandSessionID() []byte {
    buf := make([]byte, SessionIDLength)
    rand.Read(buf)
    return buf
}

func handleConnRead(id []byte, conn net.Conn, sendChan chan *DataBlock, errChan chan error) {
    for {
        buf := make([]byte, DefaultBlockLength)
        n, err := conn.Read(buf)
        if err != nil {
            errChan <- err
            return
        }
        block := &DataBlock{
            SessionID: id,
            Length: int32(n),
            Data: buf,
        }
        sendChan <- block
    }
}

func handleConnWrite(id []byte, conn net.Conn, recvChan chan *DataBlock, errChan chan error) {
    for {
        block := <-recvChan
        if block.Length == END_SESSION {
            // nil means connection closed gracefully
            errChan <- nil
            return
        } else {
            // regular connection data
            _, err := conn.Write(block.Data[:block.Length])
            if err != nil {
                errChan <- err
            }
        }
    }
}
