package dns

import (
	"fmt"
	"github.com/miekg/dns"
	"net"
	"time"
)

var servers = []string{"8.8.8.8"}

const fallback = "213.174.39.77"

func Lookup(host string) net.IP {
	ip := fromCache(host)
	if ip != nil {
		return ip
	}

	for _, server := range servers {
		msg := new(dns.Msg)
		msg.SetQuestion(host+".", dns.TypeA)
		client := new(dns.Client)
		response, _, err := client.Exchange(msg, server+":53")
		if err != nil {
			fmt.Println(err)
			continue
		}

		for _, answer := range response.Answer {
			switch record := answer.(type) {
			case *dns.A:
				ip := record.A
				toCache(host, ip, time.Duration(record.Hdr.Ttl)*time.Second)
				return ip
			}
		}
	}

	fmt.Println("Using fallback IP")
	return net.ParseIP(fallback)
}
