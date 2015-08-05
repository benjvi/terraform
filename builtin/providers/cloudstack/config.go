package cloudstack

import "github.com/benjvi/go-cloudstack/cloudstack43"

// Config is the configuration structure used to instantiate a
// new CloudStack client.
type Config struct {
	APIURL      string
	APIKey      string
	SecretKey   string
	HTTPGETOnly bool
	Timeout     int64
}

// Client() returns a new CloudStack client.
func (c *Config) NewClient() (*cloudstack43.CloudStackClient, error) {
	cs := cloudstack43.NewAsyncClient(c.ApiURL, c.ApiKey, c.SecretKey, false)
	cs.AsyncTimeout(c.Timeout)
	return cs, nil
}
