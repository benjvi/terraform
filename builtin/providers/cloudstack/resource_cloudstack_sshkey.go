package cloudstack

import (
	"fmt"
	"log"
	"strings"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/xanzy/go-cloudstack/cloudstack"
)

func resourceCloudStackSSHKey() *schema.Resource {
	return &schema.Resource{
		Create: resourceCloudStackSSHKeyCreate,
		Read:   resourceCloudStackSSHKeyRead,
		Delete: resourceCloudStackSSHKeyDelete,

		Schema: map[string]*schema.Schema{
			"name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"public_key": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},

			"private_key": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},

			"fingerprint": &schema.Schema{
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceCloudStackSSHKeyCreate(d *schema.ResourceData, meta interface{}) error {
	cs := meta.(*cloudstack.CloudStackClient)

	name := d.Get("name").(string)
	public_key, ok := d.GetOk("public_key")
	log.Printf("[DEBUG] Public key value supplied: %s", public_key)
	log.Printf("[DEBUG] Public key retrieval status: %s", ok)
	//TODO: validate format of public key
	if ok && public_key != "" {
		//register key supplied
		//todo: check the key supplied is valid
		p := cs.SSH.NewRegisterSSHKeyPairParams(name, public_key.(string))
		r, err := cs.SSH.RegisterSSHKeyPair(p)
		if err != nil {
			return err
		}
		log.Printf("[DEBUG] RegisterSSHKeyPair response: %+v\n", r)
		log.Printf("[DEBUG] Key pair successfully registered at Cloudstack")
		d.SetId(name)
	} else {
		//no key supplied, must create one and return the private key
		p := cs.SSH.NewCreateSSHKeyPairParams(name)
		r, err := cs.SSH.CreateSSHKeyPair(p)
		if err != nil {
			return err
		}
		log.Printf("[DEBUG] CreateSSHKeyPair response: %+v\n", r)
		log.Printf("[DEBUG] Key pair successfully generated at Cloudstack")
		log.Printf("[DEBUG] Private key returned: %s", r.Privatekey)
		d.Set("private_key", r.Privatekey)
		d.SetId(name)
	}

	return resourceCloudStackSSHKeyRead(d, meta)
}

func resourceCloudStackSSHKeyRead(d *schema.ResourceData, meta interface{}) error {
	cs := meta.(*cloudstack.CloudStackClient)

	log.Printf("[DEBUG] looking for ssh key  %s with name %s", d.Id(), d.Get("name").(string))
	p := cs.SSH.NewListSSHKeyPairsParams()
	p.SetName(d.Get("name").(string))
	r, err := cs.SSH.ListSSHKeyPairs(p)
	if err != nil {
		return err
	}
	if r.Count == 0 {
		//todo: check if this is ever triggered, soemtimes no results doesnt send this back
		log.Printf("[DEBUG] Key pair %s does not exist", d.Get("name").(string))
		d.Set("name", "")
		return nil
	} else {
		//SSHKeyPair name is unique in a cloudstack account so dont need to check for multiple
		d.Set("name", r.SSHKeyPairs[0].Name)
		d.Set("fingerprint", r.SSHKeyPairs[0].Fingerprint)
		log.Printf("[DEBUG] Read ssh key pair %+v\n", d)
	}

	return nil
}

func resourceCloudStackSSHKeyDelete(d *schema.ResourceData, meta interface{}) error {
	cs := meta.(*cloudstack.CloudStackClient)

	// Create a new parameter struct
	p := cs.SSH.NewDeleteSSHKeyPairParams(d.Get("name").(string))

	// Remove the NIC
	_, err := cs.SSH.DeleteSSHKeyPair(p)
	if err != nil {
		// This is a very poor way to be told the UUID does no longer exist :(
		if strings.Contains(err.Error(), fmt.Sprintf(
			"A key pair with name '%s' does not exist for account", d.Get("name").(string))) {
			return nil
		}

		return fmt.Errorf("Error deleting SSH Keypair: %s", err)
	}

	return nil
}
