package netapi

import (
	"fmt"
	"log"
	"time"

	"github.com/benjvi/go-net-api"
	"github.com/hashicorp/terraform/helper/schema"
)

func resourceNetAPIPrivateDirectConnect() *schema.Resource {
	return &schema.Resource{
		Create: resourceNetAPIPrivateDirectConnectCreate,
		Read:   resourceNetAPIPrivateDirectConnectRead,
		Update: resourceNetAPIPrivateDirectConnectUpdate,
		Delete: resourceNetAPIPrivateDirectConnectDelete,

		Schema: map[string]*schema.Schema{
			"cidr": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},

			"dcg": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"display_text": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},

			"gateway": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},

			"zone": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"region": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"sid": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},

			"router_endpoint_one": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},

			"router_endpoint_two": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},

			"vlan": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},

			"is_ready": &schema.Schema{
				Type:     schema.TypeBool,
				Computed: true,
			},

			"is_ready_timeout": &schema.Schema{
				Type:     schema.TypeInt,
				Optional: true,
				Default:  3000,
			},
		},
	}
}

func resourceNetAPIPrivateDirectConnectCreate(d *schema.ResourceData, meta interface{}) error {
	cs := meta.(*netAPI.NetAPIClient)

	displaytext := d.Get("display_text").(string)
	region := d.Get("region").(string)

	p := cs.Network.NewListNetworksParams(region)
	p.SetSubtype("privatedirectconnect")
	r, err := cs.Network.ListNetworks(p)
	if err != nil {
		return fmt.Errorf("Error checking for existing private direct connect %s: %s", displaytext, err)
	}

	// Can't search directly for displaytext, so we must filter the list of all VPNs
	fr := make([]*netAPI.Network, 0, len(r.Networks))
	for _, val := range r.Networks {
		if val.Displaytext == displaytext {
			fr = append(fr, val)
		}
	}

	if len(fr) == 0 {
		// Create the private direct connect
		zonename := d.Get("zone").(string)
		cidr := d.Get("cidr").(string)
		gateway := d.Get("gateway").(string)

		// Create a new parameter struct
		p2 := cs.PrivateDirectConnect.NewCreatePrivateDirectConnectParams(displaytext, zonename, cidr, gateway, region)
		p2.SetDcgname(d.Get("dcg").(string))

		// Create the new private direct connect
		r2, err := cs.PrivateDirectConnect.CreatePrivateDirectConnect(p2)
		if err != nil {
			return fmt.Errorf("Error creating private direct connect %s: %s", displaytext, err)
		}

		d.SetId(r2.ListPrivateDirectConnects[0].Id)
		d.Set("sid", r2.ListPrivateDirectConnects[0].Sid)
		d.Set("router_endpoint_one", r2.ListPrivateDirectConnects[0].Routerendpoint1)
		d.Set("router_endpoint_two", r2.ListPrivateDirectConnects[0].Routerendpoint2)
		d.Set("vlan", r2.ListPrivateDirectConnects[0].Vlan)
	} else if len(fr) == 1 {
		// Network already exists so we must adopt it
		d.SetId(fr[0].Id)

		// unfortunately we can now never get hold of the computed values
		d.Set("sid", "VALUE_UNAVAILABLE")
		d.Set("router_endpoint_one", "VALUE_UNAVAILABLE")
		d.Set("router_endpoint_two", "VALUE_UNAVAILABLE")
		d.Set("vlan", "VALUE_UNAVAILABLE")
	} else {
		return fmt.Errorf("You have multiple private direct connects with the same identifier (%s)", displaytext)
	}

	// Wait until the direct connect is ready to use, or timeout with an error...
	currentTime := time.Now().Unix()
	timeout := int64(d.Get("is_ready_timeout").(int))
	for {
		// Start with the sleep so the provision action has a few seconds
		// to be processed correctly
		time.Sleep(10 * time.Second)

		err := resourceNetAPIPrivateDirectConnectRead(d, meta)
		if err != nil {
			return err
		}

		if d.Get("is_ready").(bool) {
			return nil
		}

		if time.Now().Unix()-currentTime > timeout {
			return fmt.Errorf("Timeout while waiting for private direct connect to become ready")
		}
	}
}

func resourceNetAPIPrivateDirectConnectRead(d *schema.ResourceData, meta interface{}) error {
	cs := meta.(*netAPI.NetAPIClient)

	region := d.Get("region").(string)

	n, count, err := cs.Network.GetNetworkByID(d.Id(), region)
	if err != nil {
		if count == 0 {
			log.Printf(
				"[DEBUG] Network %s does no longer exist", d.Get("display_text").(string))
			d.SetId("")
			return nil
		}

		return err
	}

	d.Set("display_text", n.Displaytext)
	d.Set("cidr", n.Cidr)
	d.Set("dcg", n.Dcgfriendlyname)
	d.Set("gateway", n.Gateway)
	d.Set("is_ready", n.Isprovisioned)
	return nil
}

func resourceNetAPIPrivateDirectConnectUpdate(d *schema.ResourceData, meta interface{}) error {
	cs := meta.(*netAPI.NetAPIClient)

	if !d.Get("is_ready").(bool) {
		return fmt.Errorf("Private direct connect is not ready, cannot perform any updates")
	}

	region := d.Get("region").(string)

	// Create a new parameter struct
	p := cs.Network.NewModifyNetworkParams(d.Id(), region)

	// Check if the name or display text is changed
	if d.HasChange("display_text") {
		p.SetDisplaytext(d.Get("display_text").(string))
	}

	// Check if the cidr is changed
	// TODO: if cidr changes then gateway must also always change??
	if d.HasChange("cidr") {
		p.SetCidr(d.Get("cidr").(string))
	}

	// Check if the gateway is changed
	if d.HasChange("gateway") {
		p.SetGateway(d.Get("gateway").(string))
	}

	// Update the network
	r, err := cs.Network.ModifyNetwork(p)
	if err != nil {
		return fmt.Errorf(
			"Error updating network %s: %s", d.Get("display_text").(string), err)
	}

	//Some changes will cause resource to be created
	if r.ListNetworks[0].Id != "" {
		d.SetId(r.ListNetworks[0].Id)
	}

	// Update can make the network not ready, must wait again
	currentTime := time.Now().Unix()
	timeout := int64(d.Get("is_ready_timeout").(int))
	for {
		// Start with the sleep so the provision action has a few seconds
		// to be processed correctly
		time.Sleep(10 * time.Second)

		err := resourceNetAPIPrivateDirectConnectRead(d, meta)
		if err != nil {
			return err
		}

		if d.Get("is_ready").(bool) {
			return nil
		}

		if time.Now().Unix()-currentTime > timeout {
			return fmt.Errorf("Timeout while waiting for private direct connect to become ready")
		}
	}
}

func resourceNetAPIPrivateDirectConnectDelete(d *schema.ResourceData, meta interface{}) error {
	// Can't delete these resources so just stop managing them
	return nil
}
