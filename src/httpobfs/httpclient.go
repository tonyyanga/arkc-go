package httpobfs

import (
    "encoding/binary"
    "net/http"
    "bytes"
    "time"
    "sync"

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
    mux sync.RWMutex
}

// HTTP payload structure:
// ------------------------------
// | Session ID | Length | Data |
// ------------------------------

func (c *HTTPClient) connect(id []byte, sendChan chan *DataBlock) {
    sessionEnded := false // whether END_SESSION is issued
    var nextPollInterval time.Duration = 0 // Polling interval
    for {
        if sessionEnded {
            break
        }

        var req *http.Request

        buf := &bytes.Buffer{}

        // construct HTTP request
        select {
        case block := <-sendChan:
            // Have data to send out
            // First, allocate buffer for data to send out
            if block.Length == END_SESSION {
                sessionEnded = true
            }

            // Write data to buffer
            buf.Write(block.SessionID)
            err := binary.Write(buf, binary.BigEndian, block.Length)
            if err != nil {
                panic("Error when writing DataBlock info to buffer")
            }

            if block.Length > 0 {
                buf.Write(block.Data[:block.Length])
            }

            // Create req object
            req, err = http.NewRequest("POST", c.url, buf)
            if err != nil {
                panic("Error when creating http request object")
            }
        case <- time.After(nextPollInterval):
            var err error
            buf.Write(id)
            err = binary.Write(buf, binary.BigEndian, NO_DATA)
            if err != nil {
                panic("Error when writing DataBlock info to buffer")
            }

            req, err = http.NewRequest("POST", c.url, buf)
            if err != nil {
                panic("Error when creating http request object")
            }
        }

        if nextPollInterval == 0 {
            nextPollInterval = minPollInterval
        } else {
            nextPollInterval = 2 * nextPollInterval
            if nextPollInterval > maxPollInterval {
                nextPollInterval = maxPollInterval
            }
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

        if resp.ContentLength > 0 || resp.ContentLength == -1 {
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
                c.mux.RLock()
                _, exists := c.chanMap[string(block.SessionID)]
                c.mux.RUnlock()
                if !exists {
                    log.Printf("Error: Data block with unknown session id received")
                    continue
                }
            }

            c.RecvChan <- block

            nextPollInterval = 0 // immediate poll since data is returned
        }
    }
}

// Start HTTP polling client with the http.Client provided
func (c *HTTPClient) StartWithHTTPClient(url string, client *http.Client) {
    c.url = url
    c.client = client

    c.chanMap = make(map[string] chan *DataBlock)

    // Every goroutine only polls on one connection to avoid reordering
    for {
        block := <-c.SendChan
        if block.Length == NEW_SESSION {
            targetChan := make(chan *DataBlock, 10)
            c.mux.Lock()
            c.chanMap[string(block.SessionID)] = targetChan
            c.mux.Unlock()
            go c.connect(block.SessionID, targetChan)
        }

        c.mux.RLock()
        targetChan, exists := c.chanMap[string(block.SessionID)]
        c.mux.RUnlock()
        if !exists {
            panic("Attempt to send data without passing NEW_SESSION flag")
        }

        targetChan <- block

        if block.Length == END_SESSION {
            c.mux.Lock()
            delete(c.chanMap, string(block.SessionID))
            c.mux.Unlock()
        }
    }
}

// Start HTTP polling client with default http.Client
func (c *HTTPClient) Start(url string) {
    c.StartWithHTTPClient(url, &http.Client{})
}
