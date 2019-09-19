package sgulreg

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
)

// ClientConfig is the SgulREG API client configuration.
type ClientConfig struct {
	Host       string
	Scheme     string
	httpClient *http.Client
}

// Client is the SgulREG API client.
type Client struct {
	url        string
	httpClient *http.Client
}

// NewClient returns a new instance of the SgulREG API client.
func NewClient(cc ClientConfig) *Client {
	hc := http.DefaultClient
	if cc.httpClient != nil {
		hc = cc.httpClient
	}
	return &Client{
		url:        cc.Scheme + "://" + cc.Host + "/sgulreg/services",
		httpClient: hc,
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
