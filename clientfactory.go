package webauthclient

import (
	"errors"
	"net"
	"net/http"
	"time"
)

var (
	// ErrNoAddress is returned if the interface has no suitable address.
	ErrNoAddress = errors.New("webauthclient: No suitable address on given interface")
)

// HTTPClientFactoryFunc is a function that returns a new *http.Client when called.
type HTTPClientFactoryFunc func() *http.Client

// getSourceAddr returns the first IPv4 address configured on netInterface.
func getSourceAddr(netInterface string) (net.Addr, error) {
	intf, err := net.InterfaceByName(netInterface)
	if err != nil {
		return nil, err
	}
	addrs, err := intf.Addrs()
	if err != nil {
		return nil, err
	}
	for _, addr := range addrs {
		s, _, err := net.ParseCIDR(addr.String())
		if err != nil {
			continue
		}
		if s.To4() != nil {
			return &net.TCPAddr{
				IP:   s,
				Port: 0,
			}, nil
		}
	}
	return nil, ErrNoAddress
}

// ClientFactoryForInterface returns a http.Client factory that connects
// from the given interface.
func ClientFactoryForInterface(netInterface string) (HTTPClientFactoryFunc, error) {
	sourceAddr, err := getSourceAddr(netInterface)
	if err != nil {
		return nil, err
	}
	factory := func() *http.Client {
		return &http.Client{
			Transport: &http.Transport{
				DialContext: (&net.Dialer{
					Timeout:   30 * time.Second,
					KeepAlive: 30 * time.Second,
					DualStack: false,
					LocalAddr: sourceAddr,
				}).DialContext,
				MaxIdleConns:          100,
				IdleConnTimeout:       90 * time.Second,
				TLSHandshakeTimeout:   10 * time.Second,
				ExpectContinueTimeout: 1 * time.Second,
			},
		}
	}
	return factory, err
}
