package httpobfs

import (
    "net/http"
    "encoding/binary"
    "log"
)

// An HTTP server handles incoming HTTP/HTTPS connections, and interact with other
// components with DataBlock channels
type HTTPServer struct {
    // User should handle RecvChan and SendChan DataBlocks
    RecvChan chan *DataBlock
    SendChan chan *DataBlock

    // maps session id to send channels
    chanMap map[string] chan *DataBlock
}

// HTTP payload structure:
// ------------------------------
// | Session ID | Length | Data |
// ------------------------------

func (s *HTTPServer) serveHTTPPost(w http.ResponseWriter, req *http.Request) {
    var sessionID string
    // Handle request
    if req.ContentLength != 0 {
        block, err := constructDataBlock(req.Body)
        if err != nil {
            log.Printf("Error when reading HTTP request from %v: %v\n", err, req.RemoteAddr)
        }

        if block.Length == NEW_SESSION {
            sendChan := make(chan *DataBlock, 10)
            s.chanMap[string(block.SessionID)] = sendChan
        }

        s.RecvChan <- block
        sessionID = string(block.SessionID)

        if block.Length == END_SESSION {
            delete(s.chanMap, sessionID)
            return
        }
    }

    sendChan, exists := s.chanMap[sessionID]
    if !exists {
        log.Printf("Error: cannot find channel by session id")
        return
    }

    // Generate response
    select {
    case block := <-sendChan:
        // Successfully get data to send as response
        n, err := w.Write(block.SessionID)
        if n < SessionIDLength || err != nil {
            log.Printf("Error when writing session id to HTTP resp to %v\n", req.RemoteAddr)
            sendChan <- block
            return
        }

        err = binary.Write(w, binary.BigEndian, block.Length)
        if err != nil {
            log.Printf("Error when writing payload length to HTTP resp to %v\n", req.RemoteAddr)
            sendChan <- block
            return
        }

        if block.Length > 0 {
            // Finally read Data
            n, err = w.Write(block.Data)
            if n < block.Length || err != nil {
                log.Printf("Error when writing data to HTTP resp to %v\n", req.RemoteAddr)
                sendChan <- block
                return
            }
        } else if block.Length == END_SESSION {
            delete(s.chanMap, sessionID)
        }
    default:
        w.Header().Set("Content-Length", "0")
    }
}

// A simple function to handle GET requests
func serveHTTPGet(w http.ResponseWriter, req *http.Request) {
    // TODO: write some fake contents
}

// Server implements the http.Handler interface to handle requests
func (s *HTTPServer) ServeHTTP(w http.ResponseWriter, req *http.Request) {
    // Follow easy path with HTTP GET
    if req.Method == http.MethodGet {
        serveHTTPGet(w, req)
    } else if req.Method == http.MethodPost {
        s.serveHTTPPost(w, req)
    } else {
        log.Printf("Error when handling HTTP request: unexpected HTTP method, %v %v from %v\n",
                   req.Method, req.URL, req.RemoteAddr)
    }
}

func (s *HTTPServer) dispatch() {
    for {
        block := <-s.SendChan
        targetChan, exists := s.chanMap[string(block.SessionID)]
        if !exists {
            log.Printf("Error: cannot find channel by session id")
            continue
        }
        targetChan <- block
    }
}

// Call this function to listen for HTTP request
func (s *HTTPServer) ListenAndServe(listenAddr string) error {
    go s.dispatch()
    err := http.ListenAndServe(listenAddr, s)
    log.Printf("Error occurred when listening for HTTP input: %v\n", err)
    return err
}

// Call this function to listen for HTTPS request
func (s *HTTPServer) ListenAndServeTLS(listenAddr, certPath, keyPath string) error {
    go s.dispatch()
    err := http.ListenAndServeTLS(listenAddr, certPath, keyPath, s)
    log.Printf("Error occurred when listening for HTTPS input: %v\n", err)
    return err
}

