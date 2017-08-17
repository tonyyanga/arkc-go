package httpobfs

import (
    "net/http"
    "encoding/binary"
    "log"
    "sync"

//    "time"
)

// An HTTP server handles incoming HTTP/HTTPS connections, and interact with other
// components with DataBlock channels
type HTTPServer struct {
    // User should handle RecvChan and SendChan DataBlocks
    RecvChan chan *DataBlock
    SendChan chan *DataBlock

    // maps session id to send channels
    chanMap map[string] chan *DataBlock
    mux sync.RWMutex
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
            http.Error(w, "Bad request.\n", http.StatusBadRequest)
            return
        }

        if block.Length == NEW_SESSION {
            //log.Printf("Server side NEW_SESSION %v\n", block.SessionID)
            sendChan := make(chan *DataBlock, 100)
            s.mux.Lock()
            _, exists := s.chanMap[string(block.SessionID)]
            if exists {
                s.mux.Unlock()
                log.Println("Error: id already exists in channel map")
                http.Error(w, "Bad request.\n", http.StatusBadRequest)
                return
            }

            s.chanMap[string(block.SessionID)] = sendChan
            s.mux.Unlock()
        }

        if block.Length != NO_DATA {
            s.RecvChan <- block
        }
        sessionID = string(block.SessionID)

        if block.Length == END_SESSION {
            s.mux.Lock()
            delete(s.chanMap, sessionID)
            s.mux.Unlock()
            return
        }
    } else {
        log.Println("HTTP Server empty req")
        http.Error(w, "Bad request.\n", http.StatusBadRequest)
        return
    }

    s.mux.RLock()
    sendChan, exists := s.chanMap[sessionID]
    s.mux.RUnlock()
    if !exists {
        log.Printf("Error: cannot find channel by session id")
        http.Error(w, "Bad request.\n", http.StatusBadRequest)
        return
    }

    // Generate response
    select {
    case block := <-sendChan:
        // Successfully get data to send as response
        n, err := w.Write(block.SessionID)
        if n != SessionIDLength || err != nil {
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
            _, err = w.Write(block.Data[:block.Length])
            if err != nil {
                log.Printf("Error when writing data to HTTP resp to %v\n", req.RemoteAddr)
                sendChan <- block
                return
            }
        } else if block.Length == END_SESSION {
            s.mux.Lock()
            delete(s.chanMap, sessionID)
            s.mux.Unlock()
        }
    default:
        w.Header().Set("Content-Length", "0")
    }
}

// A simple function to handle GET requests
func serveHTTPGet(w http.ResponseWriter, req *http.Request) {
    http.Error(w, "Bad request.\n", http.StatusBadRequest)
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
        http.Error(w, "Bad request.\n", http.StatusBadRequest)
    }
}

func (s *HTTPServer) dispatch() {
    for {
        block := <-s.SendChan
        s.mux.RLock()
        targetChan, exists := s.chanMap[string(block.SessionID)]
        s.mux.RUnlock()
        if !exists {
            log.Printf("Error: cannot find channel by session id")
            continue
        }
        select {
        case targetChan <- block:

        //case <- time.After(2 * time.Second):
        default:
            panic("TIMEOUT")
        }
    }
}

// Call this function to listen for HTTP request
func (s *HTTPServer) ListenAndServe(listenAddr string) error {
    s.chanMap = make(map[string] chan *DataBlock)
    go s.dispatch()
    err := http.ListenAndServe(listenAddr, s)
    log.Printf("Error occurred when listening for HTTP input: %v\n", err)
    return err
}

// Call this function to listen for HTTPS request
func (s *HTTPServer) ListenAndServeTLS(listenAddr, certPath, keyPath string) error {
    s.chanMap = make(map[string] chan *DataBlock)
    go s.dispatch()
    err := http.ListenAndServeTLS(listenAddr, certPath, keyPath, s)
    log.Printf("Error occurred when listening for HTTPS input: %v\n", err)
    return err
}

