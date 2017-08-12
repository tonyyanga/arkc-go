package dnshandshake

import (
    "errors"
    "net"

    "encoding/base32"
    "encoding/binary"
)

func EncodeAddr(ip net.IP, port uint16) (string, error) {
    if len(ip) != 16 {
        return "", errors.New("Not a IPv4 address")
    }

    input := make([]byte, len(ip) + binary.Size(port))
    copy(input, ip)
    binary.BigEndian.PutUint16(input[len(ip):], port)

    return base32.StdEncoding.EncodeToString(input), nil
}

func DecodeAddr(input string) (net.IP, uint16 /* port */, error) {
    if len(input) != base32.StdEncoding.EncodedLen(16 + 2) {
        return nil, 0, errors.New("Incorrect length for decoding")
    }

    data, err := base32.StdEncoding.DecodeString(input)
    if err != nil {
        return nil, 0, err
    }

    return data[:16], binary.BigEndian.Uint16(data[16:]), nil

}

