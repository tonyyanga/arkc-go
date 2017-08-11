package dnshandshake

import (
    "github.com/miekg/dns"
)

// Send handshake request with given domain to certain server, return string
// representation of answer section
func SendRequest(query string, domain string, serverAddr string) (string, error) {
    m := new(dns.Msg)
    m.SetQuestion(dns.Fqdn(query + "." + domain), dns.TypeA)

    resp, err := dns.Exchange(m, serverAddr)
    if len(resp.Answer) > 0 {
        return resp.Answer[0].String(), err
    } else {
        return "NOT FOUND", err
    }
}


