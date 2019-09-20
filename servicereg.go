package sgul

import (
	"log"

	"github.com/ITResourcesOSS/sgul/sgulreg"
)

func getServiceRegistryURL() string {
	if !IsSet("Client.ServiceRegistry.URL") {
		return sgulreg.DefaultURL
	}
	return GetConfiguration().Client.ServiceRegistry.URL
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
