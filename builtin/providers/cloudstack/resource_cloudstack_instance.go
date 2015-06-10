package cloudstack

import (
	"crypto/sha1"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"log"
	"strings"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/benjvi/go-cloudstack/cloudstack43"
)

func resourceCloudStackInstance() *schema.Resource {
	return &schema.Resource{
		Create: resourceCloudStackInstanceCreate,
		Read:   resourceCloudStackInstanceRead,
		Update: resourceCloudStackInstanceUpdate,
		Delete: resourceCloudStackInstanceDelete,

		Schema: map[string]*schema.Schema{
			"name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"display_name": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},

			"service_offering": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
			},

			"network": &schema.Schema{
				Type:     schema.TypeList,
				Optional: true,
				ForceNew: true,
				Elem: &schema.Schema{Type: schema.TypeString},
			},

			"ipaddress": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				Computed: true,
			},

			"template": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"zone": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"keypair": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
			},

			"user_data": &schema.Schema{
				Type:     schema.TypeString,
				Optional: true,
				StateFunc: func(v interface{}) string {
					switch v.(type) {
					case string:
						hash := sha1.Sum([]byte(v.(string)))
						return hex.EncodeToString(hash[:])
					default:
						return ""
					}
				},
			},

			"expunge": &schema.Schema{
				Type:     schema.TypeBool,
				Optional: true,
				Default:  false,
			},
		},
	}
}

func resourceCloudStackInstanceCreate(d *schema.ResourceData, meta interface{}) error {
	cs := meta.(*cloudstack43.CloudStackClient)
	d.Partial(true)

	// Retrieve the service_offering UUID
	serviceofferingid, e := retrieveUUID(cs, "service_offering", d.Get("service_offering").(string))
	if e != nil {
		return e.Error()
	}

	// Retrieve the zone object
	zone, _, err := cs.Zone.GetZoneByName(d.Get("zone").(string))
	if err != nil {
		return err
	}

	// Retrieve the template UUID
	templateid, e := retrieveTemplateUUID(cs, zone.Id, d.Get("template").(string))
	if e != nil {
		return e.Error()
	}

	// Create a new parameter struct
	p := cs.VirtualMachine.NewDeployVirtualMachineParams(serviceofferingid, templateid, zone.Id)

	// Set the name
	name := d.Get("name").(string)
	p.SetName(name)

	// Set the display name
	if displayname, ok := d.GetOk("display_name"); ok {
		p.SetDisplayname(displayname.(string))
	} else {
		p.SetDisplayname(name)
	}

	if zone.Networktype == "Advanced" {
		networkSlice := []string{}
		for _, network := range d.Get("network").([]interface{}) {
			// Retrieve the network UUID
			networkid, e := retrieveUUID(cs, "network", network.(string))
			if e != nil {
				return e.Error()
			}
			//set the default network ID
			networkSlice = append(networkSlice, networkid)
		}
		p.SetNetworkids(networkSlice)
	}

	// If there is a ipaddres supplied, add it to the parameter struct
	if ipaddres, ok := d.GetOk("ipaddress"); ok {
		p.SetIpaddress(ipaddres.(string))
	}

	if keypair, ok := d.GetOk("keypair"); ok {
		p.SetKeypair(keypair.(string))
	}

	// If the user data contains any info, it needs to be base64 encoded and
	// added to the parameter struct
	if userData, ok := d.GetOk("user_data"); ok {
		ud := base64.StdEncoding.EncodeToString([]byte(userData.(string)))
		if len(ud) > 32768 {
			return fmt.Errorf(
				"The supplied user_data contains %d bytes after encoding, "+
					"this exeeds the limit of 32768 bytes", len(ud))
		}
		p.SetUserdata(ud)
	}

	// Create the new instance
	r, err := cs.VirtualMachine.DeployVirtualMachine(p, false)
	if err != nil {
		return fmt.Errorf("Error creating the new instance %s: %s", name, err)
	}
	d.SetId(r.Id)
	d.SetPartial("id")
	d.SetPartial("expunge")	

	// Wait until the operation finished
	r, err = cs.VirtualMachine.WaitForDeployVirtualMachine(r.JobID)
	if err != nil {
                return fmt.Errorf("Error creating the new instance %s: %s", name, err)
        }

	// Set the connection info for any configured provisioners
	d.SetConnInfo(map[string]string{
		"host":     r.Nic[0].Ipaddress,
		"password": r.Password,
	})
	d.Partial(false)

	return resourceCloudStackInstanceRead(d, meta)
}

func resourceCloudStackInstanceRead(d *schema.ResourceData, meta interface{}) error {
	cs := meta.(*cloudstack43.CloudStackClient)

	// Get the virtual machine details
	vm, count, err := cs.VirtualMachine.GetVirtualMachineByID(d.Id())
	if err != nil {
		if count == 0 {
			log.Printf("[DEBUG] Instance %s does no longer exist", d.Get("name").(string))
			// Clear out all details so it's obvious the instance is gone
			d.SetId("")
			return nil
		}

		return err
	}

	// Update the config
	d.Set("name", vm.Name)
	d.Set("display_name", vm.Displayname)
	d.Set("ipaddress", vm.Nic[0].Ipaddress)
	d.Set("zone", vm.Zonename)
	//NB cloudstack sometimes sends back the wrong keypair name, so dont update it

        networks := []string{}
        for _, nic := range vm.Nic {
		networks = append(networks,nic.Networkname)
	}
	setValueOrUUID(d, "service_offering", vm.Serviceofferingname, vm.Serviceofferingid)
	setValueOrUUID(d, "template", vm.Templatename, vm.Templateid)

	return nil
}

func resourceCloudStackInstanceUpdate(d *schema.ResourceData, meta interface{}) error {
	cs := meta.(*cloudstack43.CloudStackClient)
	d.Partial(true)

	name := d.Get("name").(string)

	// Check if the display name is changed and if so, update the virtual machine
	if d.HasChange("display_name") {
		log.Printf("[DEBUG] Display name changed for %s, starting update", name)

		// Create a new parameter struct
		p := cs.VirtualMachine.NewUpdateVirtualMachineParams(d.Id())

		// Set the new display name
		p.SetDisplayname(d.Get("display_name").(string))

		// Update the display name
		_, err := cs.VirtualMachine.UpdateVirtualMachine(p)
		if err != nil {
			return fmt.Errorf(
				"Error updating the display name for instance %s: %s", name, err)
		}

		d.SetPartial("display_name")
	}

	// Attributes that require reboot to update
	if d.HasChange("service_offering") || d.HasChange("keypair") {
		// Before we can actually make these changes, the virtual machine must be stopped
		_, err := cs.VirtualMachine.StopVirtualMachine(cs.VirtualMachine.NewStopVirtualMachineParams(d.Id()), true)
		if err != nil {
			return fmt.Errorf(
				"Error stopping instance %s before making changes: %s", name, err)
		}

		// Check if the service offering is changed and if so, update the offering
		if d.HasChange("service_offering") {
			log.Printf("[DEBUG] Service offering changed for %s, starting update", name)

			// Retrieve the service_offering UUID
			serviceofferingid, e := retrieveUUID(cs, "service_offering", d.Get("service_offering").(string))
			if e != nil {
				return e.Error()
			}

			// Create a new parameter struct
			p := cs.VirtualMachine.NewChangeServiceForVirtualMachineParams(d.Id(), serviceofferingid)

			// Change the service offering
			_, err = cs.VirtualMachine.ChangeServiceForVirtualMachine(p)
			if err != nil {
				return fmt.Errorf(
					"Error changing the service offering for instance %s: %s", name, err)
			}
			d.SetPartial("service_offering")
		}

		if d.HasChange("keypair") {
			log.Printf("[DEBUG] SSH keypair changed for %s, starting update", name)

			p := cs.SSH.NewResetSSHKeyForVirtualMachineParams(d.Id(), d.Get("keypair").(string))

			// Change the ssh keypair
			_, err = cs.SSH.ResetSSHKeyForVirtualMachine(p, true)
			if err != nil {
				return fmt.Errorf(
					"Error changing the SSH keypair for instance %s: %s", name, err)
			}
			d.SetPartial("keypair")
		}

		// Start the virtual machine again
		_, err = cs.VirtualMachine.StartVirtualMachine(cs.VirtualMachine.NewStartVirtualMachineParams(d.Id()), true)
		if err != nil {
			return fmt.Errorf(
				"Error starting instance %s after making changes", name)
		}
	}

	// Simple implementation - but should we combine composite changes into one action?
	if d.HasChange("user_data") {
		// If user data is removed, don't send update (??)
		if userData, ok := d.GetOk("user_data"); ok {
			ud := base64.StdEncoding.EncodeToString([]byte(userData.(string)))
			if len(ud) > 32768 {
				return fmt.Errorf(
					"The supplied user_data contains %d bytes after encoding, "+
						"this exeeds the limit of 32768 bytes", len(ud))
			}

			log.Printf("[DEBUG] User data  changed for %s, starting update", name)

			// Create a new parameter struct
		        p := cs.VirtualMachine.NewUpdateVirtualMachineParams(d.Id())
			p.SetUserdata(ud)

			// Before we can actually change the service offering, the virtual machine must be stopped
			_, err := cs.VirtualMachine.StopVirtualMachine(cs.VirtualMachine.NewStopVirtualMachineParams(d.Id()), true)
	                if err != nil {
		                return fmt.Errorf(
	                                "Error stopping instance %s before changing user data: %s", name, err)
		        }
	               // Change the service offering
		        _, err = cs.VirtualMachine.UpdateVirtualMachine(p)
			if err != nil {
				return fmt.Errorf(
					"Error changing the user data for instance %s: %s", name, err)
	                }
		        // Start the virtual machine again
			_, err = cs.VirtualMachine.StartVirtualMachine(cs.VirtualMachine.NewStartVirtualMachineParams(d.Id()), true)
	                if err != nil {
		                return fmt.Errorf(
			                "Error starting instance %s after changing user data: %s", name, err)
	                }

		        d.SetPartial("user_data")
		}
	}

	d.Partial(false)
	return resourceCloudStackInstanceRead(d, meta)
}

func resourceCloudStackInstanceDelete(d *schema.ResourceData, meta interface{}) error {
	cs := meta.(*cloudstack43.CloudStackClient)

	// Create a new parameter struct
	p := cs.VirtualMachine.NewDestroyVirtualMachineParams(d.Id())

	if d.Get("expunge").(bool) {
		p.SetExpunge(true)
	}

	log.Printf("[INFO] Destroying instance: %s", d.Get("name").(string))
	if _, err := cs.VirtualMachine.DestroyVirtualMachine(p, true); err != nil {
		// This is a very poor way to be told the UUID does no longer exist :(
		if strings.Contains(err.Error(), fmt.Sprintf(
			"Invalid parameter id value=%s due to incorrect long value format, "+
				"or entity does not exist", d.Id())) {
			return nil
		}

		return fmt.Errorf("Error destroying instance: %s", err)
	}

	return nil
}
