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
	targetsCache    []string
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
	ServiceRegistry: ServiceRegistry{
		Type:     "sgulreg",
		URL:      "http://localhost:9687",
		Fallback: []string{},
	},
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
		targetsCache:    make([]string, 0),
		serviceRegistry: clientConf.ServiceRegistry,
		logger:          GetLogger().Sugar(),
	}

	// sham.discover()
	go sham.watchRegistry()
	return sham
}

func (sc *ShamClient) watchRegistry() {
	sc.logger.Debug("start watching service registry")
	for {
		<-time.After(2 * time.Second)
		go sc.discover()
	}
}

// Discover .
func (sc *ShamClient) discover() error {
	sc.logger.Debugf("discovering endpoints for service %s", sc.serviceName)
	// endpoints := []string{}
	response, err := sc.httpClient.Get(sc.serviceRegistry.URL + "/sgulreg/services/" + sc.serviceName)
	if err != nil {
		sc.targetsCache = sc.serviceRegistry.Fallback
		sc.logger.Errorf("Error making service discovery HTTP request: %s", err)
		sc.logger.Infof("using Fallback registry for service %s: %+v", sc.serviceName, sc.targetsCache)
		return ErrFailedDiscoveryRequest
	}
	sc.logger.Debugf("discovery response content-length: %s", response.Header.Get("Content-length"))

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		sc.targetsCache = sc.serviceRegistry.Fallback
		sc.logger.Errorf("Error reading service discovery HTTP response body: %s", err)
		sc.logger.Infof("using Fallback registry for service %s: %+v", sc.serviceName, sc.targetsCache)
		return ErrFailedDiscoveryResponseBody
	}
	defer response.Body.Close()

	var serviceInfo sgulreg.ServiceInfoResponse
	json.Unmarshal([]byte(body), &serviceInfo)

	if len(serviceInfo.Instances) > 0 {
		var endpoints []string
		for _, instance := range serviceInfo.Instances {
			sc.logger.Debugf("discovered service %s endpoint serviceID: %s", sc.serviceName, instance.InstanceID)
			endpoint := fmt.Sprintf("%s://%s%s", instance.Schema, instance.Host, sc.apiPath)
			endpoints = append(endpoints, endpoint)
		}

		sc.targetsCache = endpoints
		sc.logger.Infof("discovered service %s endpoints: %+v", sc.serviceName, sc.targetsCache)
	}

	if len(sc.targetsCache) == 0 {
		sc.targetsCache = sc.serviceRegistry.Fallback
		sc.logger.Infof("using Fallback registry for service %s: %+v", sc.serviceName, sc.targetsCache)
	}

	return nil
}
