package sgul

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"time"

	"github.com/ITResourcesOSS/sgul/sgulreg"

	"go.uber.org/zap"
)

// ErrFailedDiscoveryRequest is returned when a discovery request fails.
var ErrFailedDiscoveryRequest = errors.New("Error making service discovery HTTP request")

// ErrFailedDiscoveryResponseBody is returned if the discovery client is unable to read http response body.
var ErrFailedDiscoveryResponseBody = errors.New("Error reading service discovery HTTP response body")

// ShamClient defines the struct for a sham client to an http endpoint.
// The sham client is bound to an http service by its unique system discoverable name.
type ShamClient struct {
	serviceName     string
	apiPath         string
	httpClient      *http.Client
	balancer        Balancer
	TargetsCache    []string
	serviceRegistry ServiceRegistry
	logger          *zap.SugaredLogger
}

var defaultClientConfiguration = Client{
	Timeout:               120 * time.Second,
	DialerTimeout:         2 * time.Second,
	TLSHandshakeTimeout:   10 * time.Second,
	ExpectContinueTimeout: 4 * time.Second,
	ResponseHeaderTimeout: 10 * time.Second,
	Balancing:             BalancingStrategy{Strategy: RoundRobinStrategy},
	ServiceRegistry:       ServiceRegistry{Type: "sgulreg", URL: "http://localhost:9687"},
}

func clientConfiguration() Client {
	if !IsSet("Client") {
		return defaultClientConfiguration
	}
	return GetConfiguration().Client
}

func httpClient(conf Client) *http.Client {
	return &http.Client{
		Transport: &http.Transport{
			DialContext: (&net.Dialer{
				Timeout: conf.DialerTimeout,
			}).DialContext,
			TLSHandshakeTimeout:   conf.TLSHandshakeTimeout,
			ExpectContinueTimeout: conf.ExpectContinueTimeout,
			ResponseHeaderTimeout: conf.ResponseHeaderTimeout,
		},
		Timeout: conf.Timeout,
	}
}

// NewShamClient returns a new Sham client instance bounded to a service.
func NewShamClient(serviceName string, apiPath string) *ShamClient {
	clientConf := clientConfiguration()
	sham := &ShamClient{
		serviceName:     serviceName,
		apiPath:         apiPath,
		httpClient:      httpClient(clientConf),
		balancer:        BalancerFor(clientConf.Balancing.Strategy),
		TargetsCache:    make([]string, 0),
		serviceRegistry: clientConf.ServiceRegistry,
		logger:          GetLogger().Sugar(),
	}
	ping, err := sham.pingServiceRegistry()
	if err != nil {
		sham.logger.Errorf("service registry ping error: %s", err)
	}
	sham.logger.Debugf("service registry ping status code: %d", ping)
	sham.discover()
	return sham
}

// Discover .
func (sc *ShamClient) discover() error {
	sc.logger.Infof("discovering endpoints for service %s", sc.serviceName)
	// endpoints := []string{}
	response, err := sc.httpClient.Get(sc.serviceRegistry.URL + "/sgulreg/services/" + sc.serviceName)
	if err != nil {
		sc.logger.Errorf("Error making service discovery HTTP request: %s", err)
		return ErrFailedDiscoveryRequest
	}
	sc.logger.Debugf("discovery response content-length: %s", response.Header.Get("Content-length"))

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		sc.logger.Errorf("Error reading service discovery HTTP response body: %s", err)
		return ErrFailedDiscoveryResponseBody
	}
	defer response.Body.Close()

	var serviceInfo sgulreg.ServiceInfoResponse
	json.Unmarshal([]byte(body), &serviceInfo)

	for _, instance := range serviceInfo.Instances {
		endpoint := fmt.Sprintf("%s://%s%s", instance.Schema, instance.Host, sc.apiPath)
		// endpoints = append(endpoints, endpoint)
		sc.TargetsCache = append(sc.TargetsCache, endpoint)
	}

	sc.logger.Infof("service %s endpoints: %+v", sc.serviceName, sc.TargetsCache)
	return nil
}

func (sc *ShamClient) pingServiceRegistry() (int, error) {
	req, err := http.NewRequest("GET", sc.serviceRegistry.URL+"/health", nil)
	if err != nil {
		return 0, err
	}
	resp, err := sc.httpClient.Do(req)
	if err != nil {
		return 0, err
	}
	resp.Body.Close()
	return resp.StatusCode, nil
}
