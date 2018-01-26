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
func getSourceAddr(netInterface string) (*net.TCPAddr, error) {
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

func newHTTPClient(addr *net.TCPAddr, timeoutFactor uint64) *http.Client {
	tf := time.Duration(timeoutFactor)
	return &http.Client{
		Transport: &http.Transport{
			DialContext: (&net.Dialer{
				Timeout:   tf * 10 * time.Second,
				KeepAlive: tf * 10 * time.Second,
				DualStack: false,
				LocalAddr: addr,
			}).DialContext,
			MaxIdleConns:          100,
			IdleConnTimeout:       tf * 30 * time.Second,
			TLSHandshakeTimeout:   tf * 10 * time.Second,
			ExpectContinueTimeout: tf * time.Second,
		},
	}
}

// ClientFactoryForInterface returns a http.Client factory that connects
// from the given interface.
func ClientFactoryForInterface(netInterface string, timeoutFactor uint64) (HTTPClientFactoryFunc, error) {
	sourceAddr, err := getSourceAddr(netInterface)
	if err != nil {
		return nil, err
	}
	factory := func() *http.Client {
		return newHTTPClient(sourceAddr, timeoutFactor)
	}
	return factory, err
}

// ClientFactoryForAddress returns a http.Client factory that connects from a given address.
func ClientFactoryForAddress(address net.IP, timeoutFactor uint64) HTTPClientFactoryFunc {
	factory := func() *http.Client {
		return newHTTPClient(&net.TCPAddr{
			IP:   address,
			Port: 0,
		}, timeoutFactor)
	}
	return factory
}
