package httpobfs

import (
    "net/http"
    "encoding/binary"
    "log"
    "sync"
)

// An HTTP server handles incoming HTTP/HTTPS connections, and interact with other
// components with DataBlock channels
type HTTPServer struct {
    // maps session id to send / recv channels
    ChanMap map[string] chanPair
    mux sync.RWMutex

    // Callback for new connections provided by user
    // Spec: 1. callback does not need to handle ID register / unregister
    //       2. callback does not need to process NEW_SESSION
    //       3. callback should process END_SESSION and issues it on error
    ConnHandler func(id string)
}

// Register id with HTTPServer
func (s *HTTPServer) RegisterID(id string) {
    s.mux.Lock()
    defer s.mux.Unlock()
    _, exists := s.ChanMap[id]
    if exists {
        panic("Error: adding session id that already exists")
    }

    pair := chanPair{
        RecvChan: make(chan *DataBlock),
        SendChan: make(chan *DataBlock),
    }
    s.ChanMap[id] = pair

    // Start callback
    go s.ConnHandler(id)
}

// Unregister ID with HTTP Client
func (s *HTTPServer) UnregisterID(id string) {
    s.mux.Lock()
    defer s.mux.Unlock()
    _, exists := s.ChanMap[id]
    if !exists {
        log.Println("Error: ID does not exist in channel map")
    } else {
        delete(s.ChanMap, id)
    }
}

// HTTP payload structure:
// ------------------------------
// | Session ID | Length | Data |
// ------------------------------

func (s *HTTPServer) serveHTTPPost(w http.ResponseWriter, req *http.Request) {
    var sessionID string
    // Handle request
    if req.ContentLength == 0 {
        log.Println("HTTP Server empty req")
        http.Error(w, "Bad request.\n", http.StatusBadRequest)
        return
    }

    block, err := constructDataBlock(req.Body)
    if err != nil {
        log.Printf("Error when reading HTTP request from %v: %v\n", err, req.RemoteAddr)
        http.Error(w, "Bad request.\n", http.StatusBadRequest)
        return
    }

    sessionID = string(block.SessionID)

    // If NEW_SESSION, register on map
    if block.Length == NEW_SESSION {
        //log.Printf("Server side NEW_SESSION %v\n", block.SessionID)
        s.RegisterID(sessionID)
    }

    // If connection is to close, unregister id after closing connection
    if block.Length == END_SESSION {
        defer s.UnregisterID(sessionID)
    }

    s.mux.RLock()
    pair, exists := s.ChanMap[sessionID]
    s.mux.RUnlock()
    if !exists {
        log.Printf("Error: cannot find channel by session id")
        http.Error(w, "Bad request.\n", http.StatusBadRequest)
        return
    }

    sendChan := pair.SendChan
    recvChan := pair.RecvChan

    if block.Length != NO_DATA && block.Length != NEW_SESSION {
        recvChan <- block
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
            defer s.UnregisterID(string(block.SessionID))
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

// Call this function to listen for HTTP request
func (s *HTTPServer) ListenAndServe(listenAddr string) error {
    s.ChanMap = make(map[string] chanPair)
    err := http.ListenAndServe(listenAddr, s)
    log.Printf("Error occurred when listening for HTTP input: %v\n", err)
    return err
}

// Call this function to listen for HTTPS request
func (s *HTTPServer) ListenAndServeTLS(listenAddr, certPath, keyPath string) error {
    s.ChanMap = make(map[string] chanPair)
    err := http.ListenAndServeTLS(listenAddr, certPath, keyPath, s)
    log.Printf("Error occurred when listening for HTTPS input: %v\n", err)
    return err
}

