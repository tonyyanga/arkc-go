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
    url string
    client *http.Client

    // ChanMap maps session id to channel for input and output
    ChanMap map[string] chanPair
    mux sync.RWMutex
}

// Register ID with HTTP Client
func (c *HTTPClient) RegisterID(id string) chanPair {
    c.mux.Lock()
    defer c.mux.Unlock()
    pair := chanPair{
        RecvChan: make(chan *DataBlock),
        SendChan: make(chan *DataBlock),
    }
    _, exists := c.ChanMap[id]
    if exists {
        panic("Error: ID already exists in channel map")
    }

    c.ChanMap[id] = pair

    // Start goroutine to handle it
    go c.connect([]byte(id))

    return pair
}

// Unregister ID with HTTP Client
func (c *HTTPClient) UnregisterID(id string) {
    c.mux.Lock()
    defer c.mux.Unlock()
    _, exists := c.ChanMap[id]
    if !exists {
        log.Println("Error: ID does not exist in channel map")
    } else {
        delete(c.ChanMap, id)
    }
}

// HTTP payload structure:
// ------------------------------
// | Session ID | Length | Data |
// ------------------------------

func (c *HTTPClient) connect(id []byte) {
    sessionEnded := false // whether END_SESSION is issued
    var nextPollInterval time.Duration = minPollInterval // Polling interval
    var err error

    c.mux.RLock()
    pair, ok := c.ChanMap[string(id)]
    c.mux.RUnlock()
    if !ok {
        panic("key not found")
    }
    sendChan := pair.SendChan
    recvChan := pair.RecvChan

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

        // Update poll interval
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

            // Issue END_SESSION to recvChan and return
            block := &DataBlock{
                SessionID: id,
                Length: END_SESSION,
            }
            recvChan <- block
            return
        }

        if resp.StatusCode != 200 {
            log.Printf("Error occurred: roundtrip resp status code is not OK: %v\n", resp.StatusCode)

            // Issue END_SESSION to recvChan and return
            block := &DataBlock{
                SessionID: id,
                Length: END_SESSION,
            }
            recvChan <- block
            resp.Body.Close()
            return
        }

        // Process resp
        if resp.ContentLength > 0 || resp.ContentLength == -1 {
            // Not empty response
            block, err := constructDataBlock(resp.Body)
            resp.Body.Close()
            if err != nil {
                log.Printf("Error when reading from HTTP response: %v\n", err)
            } else {

                if block.Length == END_SESSION {
                    sessionEnded = true
                } else {
                    // verify the session id exists in ChanMap
                    // server side is not supposed to create new connections
                    c.mux.RLock()
                    _, exists := c.ChanMap[string(block.SessionID)]
                    c.mux.RUnlock()
                    if !exists {
                        log.Printf("Error: Data block with unknown session id received. %#X\n", block.SessionID)
                        continue
                    }
                }

                recvChan <- block

                nextPollInterval = 0 // immediate poll since data is returned
            }
        } else {
            resp.Body.Close()
        }
    }
}

// Prepare HTTP polling client with the http.Client provided
func (c *HTTPClient) StartWithHTTPClient(url string, client *http.Client) {
    c.url = url
    c.client = client

    c.ChanMap = make(map[string] chanPair)
}

// Prepare HTTP polling client with default http.Client
func (c *HTTPClient) Start(url string) {
    c.StartWithHTTPClient(url, &http.Client{})
}
