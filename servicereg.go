package sgul

import (
	"log"

	"github.com/ITResourcesOSS/sgul/sgulreg"
)

// REGAgent is the Agent used by a service to register its instance
// with the SgulREG Service Registry.
// It is an helper agent to use the sgulreg client.
type REGAgent struct {
	client *sgulreg.Client
}

func getServiceRegistryURL() string {
	if !IsSet("Client.ServiceRegistry.URL") {
		return sgulreg.DefaultURL
	}
	return GetConfiguration().Client.ServiceRegistry.URL
}

// NewREGAgent returns a new REGAgent instance
func NewREGAgent(registerURL string) *REGAgent {
	if registerURL == "" {
		registerURL = getServiceRegistryURL()
	}
	return &REGAgent{
		client: sgulreg.NewClient(registerURL),
	}
}

// Register try and register register a service with the SgulREG service.
// If the registration fails, it starts a watcher to continue trying registration.
func (ra *REGAgent) Register(r sgulreg.ServiceRegistrationRequest) (sgulreg.ServiceRegistrationResponse, error) {
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
func RegisterService(r sgulreg.ServiceRegistrationRequest) (sgulreg.ServiceRegistrationResponse, error) {
	regClient := sgulreg.NewClient(getServiceRegistryURL())
	regClient.NewRequest(r)

	response, err := regClient.Register()
	if err != nil {
		log.Printf("service registration failed: %s", err)
		log.Print("keep trying registration")
		go regClient.WatchRegistry()
	}

	return response, err
}
