package httpobfs

import (
    "encoding/binary"
    "net/http"
    "bytes"
    "time"

    "log"
)

const (
    minPollInterval = 100 * time.Millisecond
    maxPollInterval = 1 * time.Second
)

// An HTTP client does polling via HTTP/HTTPS connections, and interact with other
// components with DataBlock channels
type HTTPClient struct {
    // User should handle RecvChan and SendChan DataBlocks
    RecvChan chan *DataBlock
    SendChan chan *DataBlock

    url string
    client *http.Client

    // chanMap maps session id to channel for input
    chanMap map[string] chan *DataBlock
}

// HTTP payload structure:
// ------------------------------
// | Session ID | Length | Data |
// ------------------------------

func (c *HTTPClient) connect(sendChan chan *DataBlock) {
    sessionEnded := false // whether END_SESSION is issued
    var nextPollInterval time.Duration = 0 // Polling interval
    for {
        if sessionEnded {
            break
        }

        var req *http.Request

        // construct HTTP request
        select {
        case block := <-sendChan:
            // Have data to send out
            // First, allocate buffer for data to send out
            bufLength := SessionIDLength + binary.Size(block.Length)
            if block.Length > 0 {
                bufLength += block.Length
            } else if block.Length == END_SESSION {
                sessionEnded = true
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
        case <- time.After(nextPollInterval):
            req, err := http.NewRequest("POST", c.url, nil)
            if err != nil {
                panic("Error when creating http request object")
            }
            req.Header.Set("Content-Length", "0")
        }

        // RoundTrip
        resp, err := c.client.Do(req)
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

            if block.Length == END_SESSION {
                sessionEnded = true
            } else {
                // verify the session id exists in chanMap
                // server side is not supposed to create new connections
                _, exists := c.chanMap[string(block.SessionID)]
                if !exists {
                    log.Printf("Error: Data block with unknown session id received")
                    continue
                }
            }

            c.RecvChan <- block

            nextPollInterval = 0 // immediate poll since data is returned
        } else {
            if nextPollInterval == 0 {
                nextPollInterval = minPollInterval
            } else {
                nextPollInterval = 2 * nextPollInterval
                if nextPollInterval > maxPollInterval {
                    nextPollInterval = maxPollInterval
                }
            }
        }
    }
}

// Start HTTP polling client with the http.Client provided
func (c *HTTPClient) StartWithHTTPClient(url string, client *http.Client) error {
    c.url = url
    c.client = client

    c.chanMap = make(map[string] chan *DataBlock)

    // Every goroutine only polls on one connection to avoid reordering
    for {
        block := <-c.SendChan
        if block.Length == NEW_SESSION {
            targetChan := make(chan *DataBlock, 10)
            c.chanMap[string(block.SessionID)] = targetChan
            go c.connect(targetChan)
        }

        targetChan, exists := c.chanMap[string(block.SessionID)]
        if !exists {
            panic("Attempt to send data without passing NEW_SESSION flag")
        }

        targetChan <- block

        if block.Length == END_SESSION {
            delete(c.chanMap, string(block.SessionID))
        }
    }
    return nil
}

// Start HTTP polling client with default http.Client
func (c *HTTPClient) Start(url string) error {
    return c.StartWithHTTPClient(url, &http.Client{})
}
