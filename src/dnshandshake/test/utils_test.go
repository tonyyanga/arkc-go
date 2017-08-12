package dnshandshake_test

import (
    "testing"
    "math/rand"
    "bytes"

    "dnshandshake"
)

func TestAddrEncoding(t *testing.T) {
    for i := 0; i < 100; i++ {
        ip := make([]byte, 16)
        rand.Read(ip)

        var port uint16 = uint16(rand.Intn(65534) + 1)

        encoded, err := dnshandshake.EncodeAddr(ip, port)
        if err != nil {
            t.Error("ERROR when encoding")
            return
        }

        newip, newport, err := dnshandshake.DecodeAddr(encoded)
        if err != nil {
            t.Error("ERROR when decoding")
            return
        }

        if bytes.Compare(ip, newip) != 0 || port != newport {
            t.Error("Wrong encoding / decoding")
            return
        }
    }
}

