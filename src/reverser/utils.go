package reverser

import (
    "io"
    "sync"
    "time"
)

// Do io.Copy from src to dst, and return err to errChan channel
func connCopy(src io.Reader, dst io.Writer, errChan chan error) {
    _, err := io.Copy(dst, src)
    errChan <- err
}

//============================================================================
// Denotes the state of reverser client / server, i.e. how many connections active
type connState struct {
    clientCount byte
    updateChan chan byte
    mux sync.Mutex
}

// Increase the client count by one and issue updates
func (state *connState) AddClientCount() {
    state.mux.Lock()
    state.clientCount++
    state.mux.Unlock()
    state.updateChan <- state.clientCount
}

// Decrease the client count by one and issue updates
func (state *connState) ReduceClientCount() {
    state.mux.Lock()
    state.clientCount--
    state.mux.Unlock()
    state.updateChan <- state.clientCount
}

// Set the client count to num and issue updates if applicable
func (state *connState) UpdateClientCount(num byte) {
    if state.clientCount != num {
        state.mux.Lock()
        state.clientCount = num
        state.mux.Unlock()
        state.updateChan <- state.clientCount
    }
}

// Block for update on client count, and return client count
func (state *connState) WaitForUpdate() byte {
    select {
    case count := <-state.updateChan:
        return count
    case <- time.After(1 * time.Second): // TODO: avoid magic number
        return state.clientCount
    }
}

// Return raw channel that receives client count updates
func (state *connState) GetUpdateChan() chan byte {
    return state.updateChan
}
//===========================================================================

