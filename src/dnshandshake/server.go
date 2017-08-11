package dnshandshake

import (
    "log"
    "strings"
    "crypto/rand"
    "github.com/miekg/dns"
)

var recvChan chan string
var dom string // e.g. "google.com"

func handleDNS(w dns.ResponseWriter, r *dns.Msg) {
    resp := new(dns.Msg)
    resp.SetReply(r)

    // TODO: keep log of remote addr

    if r.Question[0].Qtype == dns.TypeA {
        // correct request type
        // return random IP and send query string to recvChan
        ip := make([]byte, 4)
        n, err := rand.Read(ip)
        if n != 4 || err != nil {
            panic("Unexpected error when generating response")
        }
        record := &dns.A{
            Hdr: dns.RR_Header{
                Name: r.Question[0].Name,
                Rrtype: dns.TypeA,
                Class: dns.ClassINET,
                Ttl: 0,
            },
            A: ip,
        }

        resp.Answer = append(resp.Answer, record)
        recvChan <- strings.Trim(r.Question[0].Name[:len(r.Question[0].Name) - len(dom) - 1], ". ")
    }
    w.WriteMsg(resp)
}

func Serve(domain string, // requests must be based on this domain
           net, listenAddr string, // Address for DNS server to listen for
           recv chan string, // channel where dns input is sent into
       ) error {
    recvChan = recv
    dom = strings.Trim(domain, ". ")

    dns.HandleFunc(domain, handleDNS)
    server := &dns.Server{
        Net: net,
        Addr: listenAddr,
        TsigSecret: nil,
    }

    err := server.ListenAndServe()
    if err != nil {
        log.Printf("Error occurred when serving DNS requests: %v\n", err)
    }
    return err
}

