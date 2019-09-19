package sgulreg

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"sync"
	"time"
)

// DefaultURL is the default SgulREG service url.
const DefaultURL = "http://localhost:9687"

// Client is the SgulREG API client.
type Client struct {
	url        string
	httpClient *http.Client
	req        ServiceRegistrationRequest
	reqMux     *sync.RWMutex
}

// NewClient returns a new instance of the SgulREG API client.
func NewClient(registryURL string) *Client {
	return &Client{
		url:        registryURL + "/sgulreg/services",
		httpClient: http.DefaultClient,
		reqMux:     &sync.RWMutex{},
	}
}

// NewRequest set the request struct to register the service.
func (c *Client) NewRequest(r ServiceRegistrationRequest) {
	c.reqMux.Lock()
	c.req = r
	c.reqMux.Unlock()
}

// Register sends a service registration request to the SgulREG service.
func (c *Client) Register() (ServiceRegistrationResponse, error) {
	c.reqMux.RLock()
	req := c.req
	c.reqMux.RUnlock()

	log.Printf("try service registration for service %s", req.Name)

	response := ServiceRegistrationResponse{}
	jsonRequest, _ := json.Marshal(req)
	resp, err := c.httpClient.Post(c.url, "application/json", bytes.NewBuffer(jsonRequest))
	if err != nil {
		return response, err
	}
	defer resp.Body.Close()

	var body []byte
	body, err = ioutil.ReadAll(resp.Body)
	json.Unmarshal(body, &response)
	log.Printf("service registered with the global service registry: %+v", response)
	return response, err
}

func (c *Client) watchRegistry() {
	for {
		<-time.After(2 * time.Second)
		go c.Register()
	}
}
