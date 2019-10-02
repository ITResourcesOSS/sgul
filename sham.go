// Copyright 2019 Luca Stasio <joshuagame@gmail.com>
// Copyright 2019 IT Resources s.r.l.
//
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

// Package sgul defines common structures and functionalities for applications.
// sham.go defines the ShamClient struct to be used as a service-to-service http load-balanced client.
package sgul

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/itross/sgul/registry"
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
	localRegistry   []string
	lrMutex         *sync.RWMutex
	serviceRegistry ServiceRegistry
	logger          *zap.SugaredLogger
}

// defaultClientConfiguration is a reasonably good default configuration for a ShamClient.
var defaultClientConfiguration = Client{
	Timeout:               120 * time.Second,
	DialerTimeout:         2 * time.Second,
	TLSHandshakeTimeout:   10 * time.Second,
	ExpectContinueTimeout: 4 * time.Second,
	ResponseHeaderTimeout: 10 * time.Second,
	Balancing:             BalancingStrategy{Strategy: RoundRobinStrategy},
	ServiceRegistry: ServiceRegistry{
		Type:          "sgulreg",
		URL:           "http://localhost:9687",
		Fallback:      []string{},
		WatchInterval: 2 * time.Second,
	},
}

// clientConfiguration returns the right client configuration.
func clientConfiguration() Client {
	if !IsSet("Client") {
		return defaultClientConfiguration
	}
	return GetConfiguration().Client
}

// httpClient initialize the internal http client structure with the incoming configuration.
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
		lrMutex:         &sync.RWMutex{},
		localRegistry:   make([]string, 0),
		serviceRegistry: clientConf.ServiceRegistry,
		logger:          GetLogger(),
	}

	// sham.discover()
	go sham.watchRegistry()
	return sham
}

// setLocalRegistry simply sets the local registry value (thread-safe).
func (sc *ShamClient) setLocalRegistry(endpoints []string) {
	sc.lrMutex.Lock()
	defer sc.lrMutex.Unlock()

	sc.localRegistry = endpoints
}

// watchRegistry keeps watching to the service registry continuously calling
// for service discovery.
func (sc *ShamClient) watchRegistry() {
	sc.logger.Debug("start watching service registry")
	for {
		//<-time.After(2 * time.Second)
		<-time.After(sc.serviceRegistry.WatchInterval)
		go sc.discover()
	}
}

// fallbackDiscovery sets up local registry to fallback information, only if the local registry
// is empty, otherwise it leaves the registry as is.
// fallbackDiscovery will be called if the system service discovery server does not return a response.
func (sc *ShamClient) fallbackDiscovery() {
	if len(sc.localRegistry) == 0 {
		sc.setLocalRegistry(sc.serviceRegistry.Fallback)
		sc.logger.Infof("using Fallback registry for service %s: %+v", sc.serviceName, sc.localRegistry)
	} else {
		sc.logger.Infof("continue using local registry for service %s: %+v", sc.serviceName, sc.localRegistry)
	}
}

// Discover gets service discovery information from the system service registry.
func (sc *ShamClient) discover() error {
	sc.logger.Debugf("discovering endpoints for service %s", sc.serviceName)
	response, err := sc.httpClient.Get(sc.serviceRegistry.URL + "/sgulreg/services/" + sc.serviceName)
	if err != nil {
		sc.logger.Errorf("Error making service discovery HTTP request: %s", err)
		sc.fallbackDiscovery()
		return ErrFailedDiscoveryRequest
	}
	sc.logger.Debugf("discovery response content-length: %s", response.Header.Get("Content-length"))

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		sc.logger.Errorf("Error reading service discovery HTTP response body: %s", err)
		sc.fallbackDiscovery()
		return ErrFailedDiscoveryResponseBody
	}
	defer response.Body.Close()

	var serviceInfo registry.ServiceInfoResponse
	json.Unmarshal([]byte(body), &serviceInfo)

	if len(serviceInfo.Instances) > 0 {
		var endpoints []string
		for _, instance := range serviceInfo.Instances {
			sc.logger.Debugf("discovered service %s endpoint serviceID: %s", sc.serviceName, instance.InstanceID)
			endpoint := fmt.Sprintf("%s://%s%s", instance.Schema, instance.Host, sc.apiPath)
			endpoints = append(endpoints, endpoint)
		}

		// sc.localRegistry = endpoints
		sc.setLocalRegistry(endpoints)
		sc.logger.Infof("discovered service %s endpoints: %+v", sc.serviceName, sc.localRegistry)
	}

	if len(sc.localRegistry) == 0 {
		// sc.localRegistry = sc.serviceRegistry.Fallback
		sc.setLocalRegistry(sc.serviceRegistry.Fallback)
		sc.logger.Infof("using Fallback registry for service %s: %+v", sc.serviceName, sc.localRegistry)
	}

	return nil
}
