package registry

import (
	"time"
)

// ServiceRegistrationRequest defines the structure sent from service in order to be registered with th registry.
type ServiceRegistrationRequest struct {
	Name           string `json:"name"`
	Host           string `json:"host"`
	Schema         string `json:"schema"`
	InfoURL        string `json:"infoUrl"`
	HealthCheckURL string `json:"healthCheckUrl"`
}

// ServiceRegistrationResponse defines the structure returned after a service instance registration.
type ServiceRegistrationResponse struct {
	InstanceID            string    `json:"instanceId"`
	RegistrationTimestamp time.Time `json:"registrationTimestamp"`
}

// ServiceInstanceInfo defines the struct for an instance of a specific service.
type ServiceInstanceInfo struct {
	InstanceID            string    `json:"instanceId"`
	Host                  string    `json:"host"`
	Schema                string    `json:"schema"`
	InfoURL               string    `json:"infoUrl"`
	HealthCheckURL        string    `json:"healthCheckUrl"`
	RegistrationTimestamp time.Time `json:"registrationTimestamp"`
	LastRefreshTimestamp  time.Time `json:"lastRefreshTimestamp"`
}

// ServiceInfoResponse defines the structure of the service instance response.
// This is the struct that clients receive to get service discovery info.
type ServiceInfoResponse struct {
	Name      string                `json:"name"`
	Instances []ServiceInstanceInfo `json:"instances"`
}
