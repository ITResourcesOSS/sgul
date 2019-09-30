// Copyright 2019 Luca Stasio <joshuagame@gmail.com>
// Copyright 2019 IT Resources s.r.l.
//
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

// Package sgul defines common structures and functionalities for applications.
// servicereg.go defines helpers for service registration using the kit/registry.Client.
package sgul

import (
	"log"

	"github.com/itross/sgul/registry"
)

// REGAgent is the Agent used by a service to register its instance
// with the SgulREG Service Registry.
// It is an helper agent to use the sgulreg client.
type REGAgent struct {
	client *registry.Client
}

func getServiceRegistryURL() string {
	if !IsSet("Client.ServiceRegistry.URL") {
		return registry.DefaultURL
	}
	return GetConfiguration().Client.ServiceRegistry.URL
}

// NewREGAgent returns a new REGAgent instance
func NewREGAgent(registerURL string) *REGAgent {
	if registerURL == "" {
		registerURL = getServiceRegistryURL()
	}
	return &REGAgent{
		client: registry.NewClient(registerURL),
	}
}

// Register try and register register a service with the SgulREG service.
// If the registration fails, it starts a watcher to continue trying registration.
func (ra *REGAgent) Register(r registry.ServiceRegistrationRequest) (registry.ServiceRegistrationResponse, error) {
	ra.client.NewRequest(r)

	response, err := ra.client.Register()
	if err != nil {
		log.Printf("service registration failed: %s", err)
		log.Print("keep trying registration")
		go ra.client.WatchRegistry()
	}

	return response, err
}

// RegisterService is an helper to register a service with the SgulREG service.
func RegisterService(r registry.ServiceRegistrationRequest) (registry.ServiceRegistrationResponse, error) {
	regClient := registry.NewClient(getServiceRegistryURL())
	regClient.NewRequest(r)

	response, err := regClient.Register()
	if err != nil {
		log.Printf("service registration failed: %s", err)
		log.Print("keep trying registration")
		go regClient.WatchRegistry()
	}

	return response, err
}
