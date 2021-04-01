package service

import "strings"

type arpatype uint8

const (
	arpaipv4 arpatype = iota
	arpaipv6
)

func arpaPTRToString(q string) string {
	t := arpaipv4

	if strings.Contains(q, `in-addr.arpa`) {
		q = strings.TrimRight(q, `.in-addr.arpa`)
	} else {
		t = arpaipv6
		q = strings.TrimRight(q, `.ip6.arpa`)
	}

	ip := strings.Split(q, `.`)
	for i, j := 0, len(ip)-1; i < j; i, j = i+1, j-1 {
		ip[i], ip[j] = ip[j], ip[i]
	}

	if t == arpaipv4 {
		return strings.Join(ip, `.`)
	} else {
		return strings.Join(ip, `:`)
	}
}
