package configcheck

import (
	"net"
	"net/netip"
	"net/url"
	"strings"
	"time"
)

// The checks below are the ones every service repeats for its settings: listen
// addresses, URLs and durations. Each reports its findings at the collector it
// is given, which should be positioned at the setting, and each treats an empty
// value as not set, which is not its business to reject. A check mirrors what
// the code that uses the setting does with it, so a config a check passes is
// one that code accepts.

// CheckListenAddress checks a listen address passed to net.Listen: it must be
// host:port, where the host may be empty to listen on every interface. An empty
// address is not set and is not checked.
func CheckListenAddress(findings *Collector, address string) {
	if address == "" {
		return
	}
	if _, port, err := net.SplitHostPort(address); err != nil {
		findings.Errorf("%q: %v", address, err)
	} else if port == "" {
		findings.Errorf("%q has no port", address)
	}
}

// IsLoopbackAddress reports whether a host:port listen address binds only the
// loopback interface. An empty host listens on every interface, so it is not
// loopback; neither is an address that does not parse.
func IsLoopbackAddress(address string) bool {
	host, _, err := net.SplitHostPort(address)
	if err != nil || host == "" {
		return false
	}
	if strings.EqualFold(host, "localhost") {
		return true
	}
	ip, err := netip.ParseAddr(host)

	return err == nil && ip.IsLoopback()
}

// ListenAddressesConflict reports whether two host:port listen addresses cannot
// both be bound: they name the same port, on the same host or with a wildcard
// host (empty, 0.0.0.0 or ::) on either side. Port 0 asks the system for a free
// port and never conflicts. Addresses that do not parse do not conflict;
// CheckListenAddress reports those.
func ListenAddressesConflict(a, b string) bool {
	hostA, portA, errA := net.SplitHostPort(a)
	hostB, portB, errB := net.SplitHostPort(b)
	if errA != nil || errB != nil || portA != portB || portA == "" || portA == "0" {
		return false
	}

	return hostA == hostB || isWildcardHost(hostA) || isWildcardHost(hostB)
}

func isWildcardHost(host string) bool {
	if host == "" {
		return true
	}
	ip, err := netip.ParseAddr(host)

	return err == nil && ip.IsUnspecified()
}

// CheckURL checks a URL a client dials: it must parse and have a host. Any
// scheme is accepted, since a client may dial any scheme with a host and port.
// An empty URL is not set and is not checked.
func CheckURL(findings *Collector, rawURL string) {
	if rawURL == "" {
		return
	}
	u, err := url.Parse(rawURL)
	if err != nil {
		findings.Errorf("%v", err)

		return
	}
	if u.Host == "" {
		findings.Errorf("%q must be a URL with a host, such as https://example.com/path", rawURL)
	}
}

// CheckHTTPURL checks a URL an HTTP client sends requests to: it must be an
// http or https URL with a host. An empty URL is not set and is not checked.
func CheckHTTPURL(findings *Collector, rawURL string) {
	if rawURL == "" {
		return
	}
	u, err := url.Parse(rawURL)
	if err != nil {
		findings.Errorf("%v", err)

		return
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		findings.Errorf("%q must be an http or https URL", rawURL)

		return
	}
	if u.Host == "" {
		findings.Errorf("%q has no host", rawURL)
	}
}

// CheckNonNegative checks that a duration is not negative.
func CheckNonNegative(findings *Collector, d time.Duration) {
	if d < 0 {
		findings.Errorf("%v must not be negative", d)
	}
}
