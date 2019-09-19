package sgulreg

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
)

// Client is the SgulREG API client.
type Client struct {
	url        string
	httpClient *http.Client
}

// NewClient returns a new instance of the SgulREG API client.
func NewClient(registryURL string) *Client {
	return &Client{
		url:        registryURL + "/sgulreg/services",
		httpClient: http.DefaultClient,
	}
}

// Register sends a service registration request to the SgulREG service.
func (c *Client) Register(r ServiceRegistrationRequest) (ServiceRegistrationResponse, error) {
	response := ServiceRegistrationResponse{}
	jsonRequest, _ := json.Marshal(r)
	resp, err := c.httpClient.Post(c.url, "application/json", bytes.NewBuffer(jsonRequest))
	if err != nil {
		return response, err
	}
	defer resp.Body.Close()

	var body []byte
	body, err = ioutil.ReadAll(resp.Body)
	json.Unmarshal(body, &response)
	return response, err
}
