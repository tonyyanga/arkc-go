package dnshandshake

import (
    "log"
    "strings"
    "math/rand"
    "github.com/miekg/dns"
)

// listen for dns request that matches given domain, and send the handshake
// query (of which domain is not included) to recv channel
func Serve(domain string, // requests must be based on this domain
           net, listenAddr string, // Address for DNS server to listen for
           recvChan chan string, // channel where dns input is sent into
       ) error {
    dom := strings.Trim(domain, ". ")

    // handleDNS is closure that preserves domain and recvChan
    handleDNS := func(w dns.ResponseWriter, r *dns.Msg) {
        resp := new(dns.Msg)
        resp.SetReply(r)

        // TODO: keep log of remote addr

        if r.Question[0].Qtype == dns.TypeA {
            // correct request type
            // return random IP and send query string to recvChan
            ip := make([]byte, 4)
            rand.Read(ip)

            // TODO: should include SOA record if this server will "act like" SOA
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
