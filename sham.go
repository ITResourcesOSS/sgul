package sgul

import (
	"net"
	"net/http"
	"time"
)

// ShamClient defines the struct for a sham client to an http endpoint.
// The sham client is bound to an http service by its unique system discoverable name.
type ShamClient struct {
	serviceName  string
	httpClient   *http.Client
	balancer     Balancer
	targetsCache []string
}

// NewShamClient returns a new Sham client instance bounded to a service.
func NewShamClient(serviceName string, lbStrategy string) *ShamClient {
	return &ShamClient{
		serviceName: serviceName,
		httpClient: &http.Client{
			Transport: &http.Transport{
				DialContext: (&net.Dialer{
					Timeout: 3 * time.Second,
				}).DialContext,
				TLSHandshakeTimeout:   10 * time.Second,
				ExpectContinueTimeout: 4 * time.Second,
				ResponseHeaderTimeout: 10 * time.Second,
			},
			Timeout: 10 * time.Minute,
		},
		balancer:     BalancerFor(lbStrategy),
		targetsCache: make([]string, 0),
	}
}
