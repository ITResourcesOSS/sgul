package sgul

import (
	"github.com/ITResourcesOSS/sgul/sgulreg"
)

func getServiceRegistryURL() string {
	l := logger.Sugar()
	if !IsSet("Client.ServiceRegistry.URL") {
		l.Debug("Client.ServiceRegistry.URL not configured. Returning default service registry url")
		return sgulreg.DefaultURL
	}
	return GetConfiguration().Client.ServiceRegistry.URL
}

// RegisterService is an helper to register a service with the SgulREG service.
func RegisterService(r sgulreg.ServiceRegistrationRequest) (sgulreg.ServiceRegistrationResponse, error) {
	regClient := sgulreg.NewClient(getServiceRegistryURL())
	regClient.NewRequest(sgulreg.ServiceRegistrationRequest{
		Name:           "test-service",
		Host:           "localhost:1111",
		Schema:         "http",
		InfoURL:        "http://localhost:1111/info",
		HealthCheckURL: "http://localhost:1111/health",
	})
	return regClient.Register()
}
