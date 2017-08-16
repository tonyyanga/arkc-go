package httpobfs

import (
    "math/rand"
    "net"

    "log"
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
        log.Printf("conn Read finished for length %v\n", n)
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
            log.Printf("conn Write finished for length %v\n", block.Length)
            if err != nil {
                errChan <- err
            }
        }
    }
}
