package sgul

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"time"

	"go.uber.org/zap"
)

// ErrFailedDiscoveryRequest is returned when a discovery request fails.
var ErrFailedDiscoveryRequest = errors.New("Error making service discovery HTTP request")

// ErrFailedDiscoveryResponseBody is returned if the discovery client is unable to read http response body.
var ErrFailedDiscoveryResponseBody = errors.New("Error reading service discovery HTTP response body")

// DiscoveryClient defines the struct for the Sham internal service Discovery client
type DiscoveryClient struct {
	serviceRegistry ServiceRegistry
	httpClient      *http.Client
}

// ServiceInstanceInfo .
type ServiceInstanceInfo struct {
	InstanceID            string    `json:"instanceId"`
	Host                  string    `json:"host"`
	Schema                string    `json:"schema"`
	InfoURL               string    `json:"infoUrl"`
	HealthCheckURL        string    `json:"healthCheckUrl"`
	RegistrationTimestamp time.Time `json:"registrationTimestamp"`
	LastRefreshTimestamp  time.Time `json:"lastRefreshTimestamp"`
}

// ServiceInfoResponse .
type ServiceInfoResponse struct {
	Name      string                `json:"name"`
	Instances []ServiceInstanceInfo `json:"instances"`
}

// ShamClient defines the struct for a sham client to an http endpoint.
// The sham client is bound to an http service by its unique system discoverable name.
type ShamClient struct {
	serviceName  string
	apiPath      string
	httpClient   *http.Client
	balancer     Balancer
	targetsCache []string
	// discovery    *DiscoveryClient
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
	ServiceRegistry:       ServiceRegistry{Type: "sgulreg", URL: "http://localhost:9687/sgulreg/services"},
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

// // NewDiscoveryClient returns a new instance of Discovery Client.
// func NewDiscoveryClient(conf Client) *DiscoveryClient {
// 	return &DiscoveryClient{
// 		serviceRegistry: conf.ServiceRegistry,
// 		httpClient:      httpClient(conf),
// 	}
// }

// NewShamClient returns a new Sham client instance bounded to a service.
func NewShamClient(serviceName string, apiPath string) *ShamClient {
	clientConf := clientConfiguration()
	sham := &ShamClient{
		serviceName:  serviceName,
		apiPath:      apiPath,
		httpClient:   httpClient(clientConf),
		balancer:     BalancerFor(clientConf.Balancing.Strategy),
		targetsCache: make([]string, 0),
		// discovery:    NewDiscoveryClient(clientConfe),
		serviceRegistry: clientConf.ServiceRegistry,
		logger:          GetLogger().Sugar(),
	}
	sham.targetsCache, _ = sham.Discover()
	return sham
}

// Discover .
func (sc *ShamClient) Discover() ([]string, error) {
	sc.logger.Infof("discovering endpoints for service %s", sc.serviceName)
	endpoints := []string{}
	response, err := sc.httpClient.Get(sc.serviceRegistry.URL + "/" + sc.serviceName)
	if err != nil {
		sc.logger.Errorf("Error making service discovery HTTP request: %s", err)
		return endpoints, ErrFailedDiscoveryRequest
	}
	sc.logger.Debugf("discovery response content-length: %s", response.Header.Get("Content-length"))

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		sc.logger.Errorf("Error reading service discovery HTTP response body: %s", err)
		return endpoints, ErrFailedDiscoveryResponseBody
	}
	defer response.Body.Close()

	var serviceInfo ServiceInfoResponse
	json.Unmarshal([]byte(body), &serviceInfo)

	for _, instance := range serviceInfo.Instances {
		endpoint := fmt.Sprintf("%s://%s%s", instance.Schema, instance.Host, sc.apiPath)
		endpoints = append(endpoints, endpoint)
	}

	sc.logger.Infof("service %s endpoints: %+v", sc.serviceName, endpoints)
	return endpoints, nil
}
