package httpobfs

import (
    "encoding/binary"
    "net/http"
    "bytes"

    "log"
)

// An HTTP client does polling via HTTP/HTTPS connections, and interact with other
// components with DataBlock channels
type HTTPClient struct {
    // User should handle RecvChan and SendChan DataBlocks
    RecvChan chan *DataBlock
    SendChan chan *DataBlock

    // Number of polling connection in parallel
    Parallelism int

    url string
    errChan chan error
}

// HTTP payload structure:
// ------------------------------
// | Session ID | Length | Data |
// ------------------------------

func (c *HTTPClient) connect() {
    client := &http.Client{}
    for {
        var req *http.Request

        // construct HTTP request
        select {
        case block := <-c.SendChan:
            // Have data to send out
            // First, allocate buffer for data to send out
            bufLength := SessionIDLength + binary.Size(block.Length)
            if block.Length > 0 {
                bufLength += block.Length
            }

            // Write data to buffer
            buf := bytes.NewBuffer(make([]byte,bufLength))
            _, err := buf.Write(block.SessionID)
            if err != nil {
                panic("Error when writing DataBlock info to buffer")
            }

            err = binary.Write(buf, binary.BigEndian, block.Length)
            if err != nil {
                panic("Error when writing DataBlock info to buffer")
            }

            if block.Length > 0 {
                n, err := buf.Write(block.Data)
                if n < block.Length || err != nil {
                    log.Println("Error occurred when building polling HTTP request: corrupted input")
                    continue
                }
            }

            // Create req object
            req, err = http.NewRequest("POST", c.url, buf)
            if err != nil {
                panic("Error when creating http request object")
            }
        //case <- time.After(1 * time.Millisecond):
        default:
            // TODO: introduce polling interval based on data successful rate
            req, err := http.NewRequest("POST", c.url, nil)
            if err != nil {
                panic("Error when creating http request object")
            }
            req.Header.Set("Content-Length", "0")
        }

        // RoundTrip
        resp, err := client.Do(req)
        if err != nil {
            log.Printf("Error occurred in HTTP client roundtrip: %v\n", err)
            continue
        }

        // Process resp
        if resp.StatusCode != 200 {
            log.Printf("Error occurred: roundtrip resp status code is not OK: %v\n", resp.StatusCode)
            continue
        }

        if resp.ContentLength > 0 {
            // Not empty response
            block, err := constructDataBlock(resp.Body)
            if err != nil {
                log.Printf("Error when reading from HTTP response: %v\n", err)
            }

            c.RecvChan <- block
        }
    }
}

// Start HTTP polling client
func (c *HTTPClient) Start(url string) error {
    c.url = url
    c.errChan = make(chan error)

    for i := 0; i < c.Parallelism; i++ {
        go c.connect();
    }

    return <- c.errChan
}
