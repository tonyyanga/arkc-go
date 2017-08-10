package httpobfs

import (
    "net/http"
    "encoding/binary"
    "time"
    "log"
)

// An HTTP server handles incoming HTTP/HTTPS connections, and interact with other
// components with DataBlock channels
type HTTPServer struct {
    // User should handle RecvChan and SendChan DataBlocks
    RecvChan chan DataBlock
    SendChan chan DataBlock
}

// HTTP payload structure:
// ------------------------------
// | Session ID | Length | Data |
// ------------------------------

func (s *HTTPServer) ServeHTTPPost(w http.ResponseWriter, req *http.Request) {
    // Handle request
    if req.ContentLength != 0 {
        // Not empty request, need to read from Body
        block := DataBlock{
            SessionID: make([]byte, SessionIDLength),
        }

        // First read Session ID from payload
        n, err := req.Body.Read(block.SessionID)
        if n < SessionIDLength || err != nil {
            log.Printf("Error when reading session id from HTTP request from %v\n", req.RemoteAddr)
            return
        }

        // Next read Data length
        err = binary.Read(req.Body, binary.BigEndian, block.Length)
        if err != nil {
            log.Printf("Error when reading payload length from HTTP request from %v\n", req.RemoteAddr)
            return
        }

        if block.Length > 0 {
            // Finally read Data
            block.Data = make([]byte, block.Length)
            n, err = req.Body.Read(block.Data)
            if n < block.Length || err != nil {
                log.Printf("Error when reading data from HTTP request from %v\n", req.RemoteAddr)
                return
            }
        }

        s.RecvChan <- block
    }

    // Generate response
    select {
    case block := <-s.SendChan:
        // Successfully get data to send as response
        n, err := w.Write(block.SessionID)
        if n < SessionIDLength || err != nil {
            log.Printf("Error when writing session id to HTTP resp to %v\n", req.RemoteAddr)
            s.SendChan <- block
            return
        }

        err = binary.Write(w, binary.BigEndian, block.Length)
        if err != nil {
            log.Printf("Error when writing payload length to HTTP resp to %v\n", req.RemoteAddr)
            s.SendChan <- block
            return
        }

        if block.Length > 0 {
            // Finally read Data
            n, err = w.Write(block.Data)
            if n < block.Length || err != nil {
                log.Printf("Error when writing data to HTTP resp to %v\n", req.RemoteAddr)
                s.SendChan <- block
                return
            }
        }
    case <- time.After(1 * time.Second):
        w.Header().Set("Content-Length", "0")
    }
}

// A simple function to handle GET requests
func ServeHTTPGet(w http.ResponseWriter, req *http.Request) {
    // TODO: write some fake contents
}

// Server implements the http.Handler interface to handle requests
func (s *HTTPServer) ServeHTTP(w http.ResponseWriter, req *http.Request) {
    // Follow easy path with HTTP GET
    if req.Method == http.MethodGet {
        ServeHTTPGet(w, req)
        return
    } else if req.Method == http.MethodPost {
        s.ServeHTTPPost(w, req)
        return
    } else {
        log.Printf("Error when handling HTTP request: unexpected HTTP method, %v %v from %v\n",
                   req.Method, req.URL, req.RemoteAddr)
        return
    }
}

// Call this function to listen for HTTP request
func (s *HTTPServer) ListenAndServe(listenAddr string) error {
    err := http.ListenAndServe(listenAddr, s)
    log.Printf("Error occurred when listening for HTTP input: %v\n", err)
    return err
}

// Call this function to listen for HTTPS request
func (s *HTTPServer) ListenAndServeTLS(listenAddr, certPath, keyPath string) error {
    err := http.ListenAndServeTLS(listenAddr, certPath, keyPath, s)
    log.Printf("Error occurred when listening for HTTPS input: %v\n", err)
    return err
}

