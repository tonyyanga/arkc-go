package dnshandshake

import (
    "github.com/miekg/dns"
)

// Send handshake request with given domain to certain server
func SendRequest(query string, domain string, serverAddr string) error {
    m := new(dns.Msg)
    m.SetQuestion(dns.Fqdn(query + "." + domain), dns.TypeA)

    _, err := dns.Exchange(m, serverAddr)
    return err
}


