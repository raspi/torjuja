package service

import (
	"encoding/json"
	"fmt"
	"github.com/alexandrevicenzi/go-sse"
	"github.com/miekg/dns"
	"github.com/raspi/torjuja/pkg/db/iface"
	"github.com/raspi/torjuja/pkg/httpapi/frontend"
	"log"
	"net"
	"net/http"
	"os"
	"path"
	"strings"
	"time"
)

type Database struct {
	FileSystem *struct {
		Path string `json:"path"`
	} `json:"fs,omitempty"`
}

type Blocked struct {
	IPv4 string `json:"ipv4"`
	IPv6 string `json:"ipv6"`
	PTR  string `json:"ptr"`
}

type Config struct {
	ApiListen       string   `json:"api"`
	ListenAddresses []string `json:"listen"`
	Blocked         Blocked  `json:"blocked"`
	TTL             uint32   `json:"ttl"`
	Forwarders      []string `json:"forwarders"`
	Database        Database `json:"database"`
}

func LoadConfig(p string) (cfg Config, err error) {
	b, err := os.ReadFile(p)
	if err != nil {
		return cfg, err
	}

	err = json.Unmarshal(b, &cfg)
	if err != nil {
		return cfg, err
	}

	cfg.Blocked.PTR = strings.ToLower(cfg.Blocked.PTR)

	if !strings.HasSuffix(cfg.Blocked.PTR, `.`) {
		return cfg, fmt.Errorf(`PTR %q is not FQDN`, cfg.Blocked.PTR)
	}

	if len(cfg.ListenAddresses) == 0 {
		return cfg, fmt.Errorf(`no DNS servers`)
	}

	if len(cfg.Forwarders) == 0 {
		return cfg, fmt.Errorf(`no DNS forwarders`)
	}

	if cfg.Database.FileSystem != nil {
		fi, err := os.Stat(cfg.Database.FileSystem.Path)
		if err != nil {
			return cfg, err
		}

		if !fi.IsDir() {
			return cfg, fmt.Errorf(`not a directory: %q`, cfg.Database.FileSystem.Path)
		}

		if !path.IsAbs(cfg.Database.FileSystem.Path) {
			return cfg, fmt.Errorf(`not absolute path: %q`, cfg.Database.FileSystem.Path)
		}
	}

	return cfg, nil
}

type StartupFailureError error

type Service struct {
	dnsListenServers  []*dns.Server
	dnsClient         dns.Client // Generic DNS client for forwarder
	forwarders        []string   // DNS query forwarders
	errch             chan error
	httpApiListenAddr string
	db                iface.Database
	bogusIPv4         net.IP // A
	bogusIPv6         net.IP // AAAA
	bogusTTL          uint32 // Seconds
	bogusPTR          string // PTR
	blockLogger       *log.Logger
	allowLogger       *log.Logger
	logger            *log.Logger
	httpfrontend      *frontend.Server
}

func New(dnsListenAddresses []string, blocked Blocked, ttl uint32, forwarders []string, db iface.Database, httpApiListen string, errch chan error) (s *Service, err error) {
	if len(dnsListenAddresses) == 0 {
		return nil, fmt.Errorf(`no DNS servers`)
	}

	if len(forwarders) == 0 {
		return nil, fmt.Errorf(`no DNS forwarders`)
	}

	if !strings.HasSuffix(blocked.PTR, `.`) {
		return nil, fmt.Errorf(`PTR %q is not FQDN`, blocked.PTR)
	}

	bogusIPv4 := net.ParseIP(blocked.IPv4)
	bogusIPv6 := net.ParseIP(blocked.IPv6)

	s = &Service{
		logger:            log.New(os.Stdout, ``, 0),
		blockLogger:       log.New(os.Stdout, `BLOCK: `, 0),
		allowLogger:       log.New(os.Stdout, `ALLOW: `, 0),
		bogusIPv4:         bogusIPv4,
		bogusIPv6:         bogusIPv6,
		bogusPTR:          blocked.PTR,
		bogusTTL:          ttl,
		dnsClient:         dns.Client{},
		forwarders:        forwarders,
		errch:             errch,
		httpApiListenAddr: httpApiListen,
		db:                db,
		httpfrontend:      frontend.New(db),
	}

	for _, dnsserver := range dnsListenAddresses {
		mux := dns.NewServeMux()
		mux.HandleFunc(`.`, s.handleDNSReq) // Catch-all

		dnssrv := &dns.Server{
			Addr:      dnsserver,
			Net:       "udp",
			Handler:   mux,
			ReusePort: true,
		}

		s.dnsListenServers = append(s.dnsListenServers, dnssrv)
	}

	return s, nil
}

func (s *Service) Listen() error {
	go func(errs chan error) {
		if err := http.ListenAndServe(s.httpApiListenAddr, s.httpfrontend.GetRouter()); err != nil {
			errs <- StartupFailureError(err)
		}
	}(s.errch)

	for _, server := range s.dnsListenServers {
		go func(srv *dns.Server, errs chan error) {
			if err := srv.ListenAndServe(); err != nil {
				errs <- StartupFailureError(err)
			}
		}(server, s.errch)
	}

	return nil
}

// getForwarder gets DNS forwarder
func (s *Service) getForwarder() string {
	// TODO
	return s.forwarders[0]
}

func (s *Service) allowLog(name string, t string) {
	s.allowLogger.Printf(`%s %s`, t, name)
}

func (s *Service) blockLog(name string, t string) {
	s.blockLogger.Printf(`%s %s`, t, name)
	s.httpfrontend.SendMessage(`/events/blocked`, sse.SimpleMessage(fmt.Sprintf(`%s %s`, t, name)))
}

// allowedA checks Service.db for allowed DNS query
func (s *Service) allowedA(name string) bool {
	allowed, err := s.db.AllowedA(name)
	if err != nil {
		s.errch <- err
		return false
	}

	return allowed
}

// allowedAAAA checks Service.db for allowed DNS query
func (s *Service) allowedAAAA(name string) bool {
	allowed, err := s.db.AllowedAAAA(name)
	if err != nil {
		s.errch <- err
		return false
	}

	return allowed
}

// allowedPTR checks Service.db for allowed DNS query
func (s *Service) allowedPTR(name string) bool {
	allowed, err := s.db.AllowedPTR(name)
	if err != nil {
		s.errch <- err
		return false
	}

	return allowed
}

// queryForwarder sends DNS queries to external resolver.
// Answers are checked against Service.db database
func (s *Service) queryForwarder(req *dns.Msg) (resp *dns.Msg, dur time.Duration, err error) {
	resp = &dns.Msg{}
	resp.SetReply(req)
	//resp.Rcode = dns.RcodeRefused

	now := time.Now()

	reply, dur, err := s.dnsClient.Exchange(req, s.getForwarder())
	if err != nil {
		s.errch <- err
		return nil, time.Now().Sub(now), fmt.Errorf(`forwarder: %w`, err)
	}

	for _, a := range reply.Answer {
		// Process DNS query answers
		hdr := a.Header()

		switch hdr.Rrtype {
		case dns.TypeCNAME: // Skip CNAME
			s.allowLog(hdr.Name+` [forwarder]`, `CNAME`)
			resp.Answer = append(resp.Answer, a)
			continue
		}

		// Allowed?
		if !s.checkAllowed(dns.Question{
			Name:   hdr.Name,
			Qtype:  hdr.Rrtype,
			Qclass: hdr.Class,
		}) {
			s.blockLog(hdr.Name+` [forwarder]`, dns.TypeToString[hdr.Rrtype])
			return nil, time.Now().Sub(now), fmt.Errorf(`forwarder: not allowed %q`, hdr.Name)
		}

		resp.Answer = append(resp.Answer, a)
	}

	return resp, time.Now().Sub(now), nil
}

// checkDnsRequest queries database Service.db for allowed DNS query
func (s *Service) checkDnsRequest(req *dns.Msg) (resp *dns.Msg, dur time.Duration, err error) {
	now := time.Now()

	resp = &dns.Msg{}
	resp.SetReply(req)
	resp.Rcode = dns.RcodeRefused

	for _, q := range req.Question {
		// Process DNS query questions

		if s.checkAllowed(q) {
			// allowed, forward to a forwarder
			s.allowLog(q.Name, dns.TypeToString[q.Qtype])
			return s.queryForwarder(&dns.Msg{
				MsgHdr: dns.MsgHdr{
					Id:               resp.Id,
					RecursionDesired: true,
				},
				Question: []dns.Question{q},
			})
		}

		s.blockLog(q.Name, dns.TypeToString[q.Qtype])

		hdr := dns.RR_Header{
			Name:   q.Name,
			Rrtype: q.Qtype,
			Class:  q.Qclass,
			Ttl:    s.bogusTTL,
		}

		// Generate blocked answer
		switch q.Qtype {
		case dns.TypeA:
			resp.Answer = append(resp.Answer, &dns.A{
				Hdr: hdr,
				A:   s.bogusIPv4,
			})

			resp.Rcode = dns.RcodeSuccess

		case dns.TypeAAAA:
			resp.Answer = append(resp.Answer, &dns.AAAA{
				Hdr:  hdr,
				AAAA: s.bogusIPv6,
			})

			resp.Rcode = dns.RcodeSuccess

		case dns.TypePTR:
			addr := net.ParseIP(arpaPTRToString(q.Name))

			if !s.checkIPAddress(addr) {
				continue
			}

			resp.Answer = append(resp.Answer, &dns.PTR{
				Hdr: hdr,
				Ptr: s.bogusPTR,
			})

			resp.Rcode = dns.RcodeSuccess

		case dns.TypeMX:
			resp.Answer = append(resp.Answer, &dns.MX{
				Hdr:        hdr,
				Preference: 0,
				Mx:         "spam.mail." + s.bogusPTR,
			})

			resp.Rcode = dns.RcodeSuccess

		case dns.TypeCNAME:
			s.logger.Printf(`cname: %q`, req.Question[0].Name)
			return s.queryForwarder(req)
		} // /switch
	} // /for

	return resp, time.Now().Sub(now), nil
}

func (s *Service) checkIPAddress(addr net.IP) bool {
	if addr.IsMulticast() {
		return false
	}

	if addr.IsLinkLocalMulticast() {
		return false
	}
	if addr.IsLoopback() {
		return false
	}
	if addr.IsInterfaceLocalMulticast() {
		return false
	}
	if addr.IsLinkLocalUnicast() {
		return false
	}
	if addr.IsUnspecified() {
		return false
	}

	return true

}

func (s *Service) checkAllowed(q dns.Question) bool {
	name := strings.ToLower(strings.TrimRight(q.Name, `.`))

	switch q.Qtype {
	case dns.TypeA:
		return s.allowedA(name)
	case dns.TypeAAAA:
		return s.allowedAAAA(name)
	case dns.TypePTR:
		name = arpaPTRToString(name)
		addr := net.ParseIP(name)

		if !s.checkIPAddress(addr) {
			return false
		}

		return true

		// TODO
		return s.allowedPTR(name)
	case dns.TypeCNAME, dns.TypeNS, dns.TypeSOA:
		return true
	default:
		return false
	}
}

// handleDNSReq handles all DNS requests and forwards them to resolver Service.checkDnsRequest
func (s *Service) handleDNSReq(w dns.ResponseWriter, req *dns.Msg) {
	reply, _, err := s.checkDnsRequest(req)
	if err != nil {
		s.errch <- err
		return
	}

	err = w.WriteMsg(reply)
	if err != nil {
		s.errch <- err
		return
	}
}
