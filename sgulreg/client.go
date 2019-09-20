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
	registered bool
}

// NewClient returns a new instance of the SgulREG API client.
func NewClient(registryURL string) *Client {
	return &Client{
		url:        registryURL + "/sgulreg/services",
		httpClient: http.DefaultClient,
		reqMux:     &sync.RWMutex{},
		registered: false,
	}
}

// NewRequest set the request struct to register the service.
func (c *Client) NewRequest(r ServiceRegistrationRequest) {
	c.reqMux.Lock()
	c.req = r
	c.reqMux.Unlock()
}

// Register sends a service registration request to the SgulREG service.
// TODO: add a channel to return results to the WatchRegistry() func.
func (c *Client) Register() (ServiceRegistrationResponse, error) {
	c.reqMux.RLock()
	req := c.req
	c.registered = false
	c.reqMux.RUnlock()

	log.Printf("*********> try service registration for service %s", req.Name)

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
	log.Printf("*********> service registered with the global service registry: %+v", response)

	c.reqMux.RLock()
	log.Print("*********> stop service registration retries")
	c.registered = true
	c.reqMux.RUnlock()

	return response, err
}

// WatchRegistry start registration retries till the registration goes well.
func (c *Client) WatchRegistry() {
	for !c.registered {
		<-time.After(2 * time.Second)
		if !c.registered {
			go c.Register()
		}
	}
}
