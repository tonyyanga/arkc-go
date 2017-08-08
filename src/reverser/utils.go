package reverser

import (
    "net"
    "io"
    "sync"
)

func connCopy(src net.Conn, dst net.Conn, errChan chan error) {
    // TODO: more logging
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

func (state *connState) AddClientCount() {
    state.mux.Lock()
    state.clientCount++
    state.mux.Unlock()
    state.updateChan <- state.clientCount
}

func (state *connState) ReduceClientCount() {
    state.mux.Lock()
    state.clientCount--
    state.mux.Unlock()
    state.updateChan <- state.clientCount
}

func (state *connState) UpdateClientCount(num byte) {
    if state.clientCount != num {
        state.mux.Lock()
        state.clientCount = num
        state.mux.Unlock()
        state.updateChan <- state.clientCount
    }
}

func (state *connState) WaitForUpdate() byte {
    return <-state.updateChan
}

func (state *connState) GetUpdateChan() chan byte {
    return state.updateChan
}
//===========================================================================

