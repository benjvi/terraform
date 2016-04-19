package netapi

import (
	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/terraform"
)

// Provider returns a terraform.ResourceProvider.
func Provider() terraform.ResourceProvider {
	return &schema.Provider{
		Schema: map[string]*schema.Schema{
			"api_url": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("NETAPI_URL", nil),
			},

			"api_key": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("NETAPI_PUBLIC_KEY", nil),
			},

			"secret_key": &schema.Schema{
				Type:        schema.TypeString,
				Required:    true,
				DefaultFunc: schema.EnvDefaultFunc("NETAPI_SECRET_KEY", nil),
			},

			"acronym": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Default:  "",
			},

			"domain_id": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Default:  "",
			},
		},

		ResourcesMap: map[string]*schema.Resource{
			"netapi_direct_connect_group":   resourceNetAPIDirectConnectGroup(),
			"netapi_private_direct_connect": resourceNetAPIPrivateDirectConnect(),
		},

		ConfigureFunc: providerConfigure,
	}
}

func providerConfigure(d *schema.ResourceData) (interface{}, error) {
	config := Config{
		ApiURL:    d.Get("api_url").(string),
		ApiKey:    d.Get("api_key").(string),
		SecretKey: d.Get("secret_key").(string),
		Acronym:   d.Get("acronym").(string),
		Domainid:  d.Get("domain_id").(string),
	}

	return config.NewClient()
}
