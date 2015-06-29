package netapi

import (
	"log"
	"strconv"

	"github.com/benjvi/go-net-api"
	"github.com/hashicorp/terraform/helper/schema"
)

func resourceNetAPIDirectConnectGroup() *schema.Resource {
	return &schema.Resource{
		Create: resourceNetAPIDirectConnectGroupCreate,
		Read:   resourceNetAPIDirectConnectGroupRead,
		Delete: resourceNetAPIDirectConnectGroupDelete,

		Schema: map[string]*schema.Schema{
			"name": &schema.Schema{
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},

			"sids": &schema.Schema{
				Type:     schema.TypeList,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Computed: true,
			},

			"networkids": &schema.Schema{
				Type:     schema.TypeList,
				Elem:     &schema.Schema{Type: schema.TypeString},
				Computed: true,
			},
		},
	}
}

func resourceNetAPIDirectConnectGroupCreate(d *schema.ResourceData, meta interface{}) error {
	cs := meta.(*netAPI.NetAPIClient)

	name := d.Get("name").(string)

	// If DCG doesn't already exist, create a new one
	p := cs.DirectConnectGroup.NewCreateDirectConnectGroupParams(name)
	r, err := cs.DirectConnectGroup.CreateDirectConnectGroup(p)
	if err != nil {
		return err
	}
	// create returns ID directly as an int so convert it for our string field
	d.SetId(strconv.FormatInt(r.ListDirectConnectGroups[0].Id, 10))

	return resourceNetAPIDirectConnectGroupRead(d, meta)
}

func resourceNetAPIDirectConnectGroupRead(d *schema.ResourceData, meta interface{}) error {
	cs := meta.(*netAPI.NetAPIClient)

	log.Printf("[DEBUG] looking for direct connect group with ID: %d", d.Id())

	p := cs.DirectConnectGroup.NewListDirectConnectGroupsParams()
	p.SetId(d.Id())

	r, err := cs.DirectConnectGroup.ListDirectConnectGroups(p)
	if err != nil {
		return err
	}
	if r.Count == 0 {
		log.Printf("[DEBUG] Direct connect group %d does not exist", d.Id())
		d.SetId("")
		return nil
	}

	//ID is unique so dont need to check for multiple
	d.Set("name", r.ListDirectConnectGroups[0].Name)
	d.Set("sids", r.ListDirectConnectGroups[0].Sids)
	d.Set("networkids", r.ListDirectConnectGroups[0].Networks)

	return nil
}

func resourceNetAPIDirectConnectGroupDelete(d *schema.ResourceData, meta interface{}) error {
	// we can't delete so just stop managing the resource in terraform

	return nil
}
