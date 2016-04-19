package netapi

import "github.com/benjvi/go-net-api"

// Config is the configuration structure used to instantiate a
// new NetAPI client.
type Config struct {
	ApiURL    string
	ApiKey    string
	SecretKey string
	Acronym   string
	Domainid  string
}

// Client() returns a new CloudStack client.
func (c *Config) NewClient() (*netAPI.NetAPIClient, error) {
	cs := netAPI.NewClient(c.ApiURL, c.ApiKey, c.SecretKey, c.Acronym, c.Domainid, false)
	return cs, nil
}
