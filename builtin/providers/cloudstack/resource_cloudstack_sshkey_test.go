package cloudstack

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
	"github.com/xanzy/go-cloudstack/cloudstack"
)

func TestAccCloudStackSSHKey_create(t *testing.T) {
	var sshkey cloudstack.SSHKeyPair

	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccCheckCloudStackSSHKeyDestroy,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccCloudStackSSHKey_create,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckCloudStackSSHKeyExists(
						"cloudstack_ssh_key.foo", &sshkey),
					testAccCheckCloudStackSSHKeyCreateAttributes(&sshkey),
				),
			},
		},
	})
}

func testAccCheckCloudStackSSHKeyExists(n string, sshkey *cloudstack.SSHKeyPair) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.Attributes["name"] == "" {
			return fmt.Errorf("No ssh key name is set")
		}

		cs := testAccProvider.Meta().(*cloudstack.CloudStackClient)
		p := cs.SSH.NewListSSHKeyPairsParams()
		p.SetName(rs.Primary.Attributes["name"])
		list, err := cs.SSH.ListSSHKeyPairs(p)

		if err != nil {
			return err
		}

		if list.Count == 1 && list.SSHKeyPairs[0].Name == rs.Primary.Attributes["name"] {
			//ssh key exists
			//set list val or n here - got inconsistency in tests
			*sshkey = *list.SSHKeyPairs[0]
			return nil
		}

		return fmt.Errorf("SSH key not found")
	}
}

func testAccCheckCloudStackSSHKeyCreateAttributes(
	sshkey *cloudstack.SSHKeyPair) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		/*if sshkey.Privatekey == "" {
			return fmt.Errorf("Empty private key")
		}*/

		if sshkey.Fingerprint == "" {
			return fmt.Errorf("Empty fingerprint")
		}
		return nil
	}
}

func testAccCheckCloudStackSSHKeyDestroy(s *terraform.State) error {
	cs := testAccProvider.Meta().(*cloudstack.CloudStackClient)

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "cloudstack_sshkey" {
			continue
		}

		if rs.Primary.Attributes["name"] == "" {
			return fmt.Errorf("No ssh key name is set")
		}

		p := cs.SSH.NewDeleteSSHKeyPairParams(rs.Primary.Attributes["name"])
		_, err := cs.SSH.DeleteSSHKeyPair(p)

		if err != nil {
			return fmt.Errorf(
				"Error deleting ssh key (%s): %s",
				rs.Primary.Attributes["name"], err)
		}
	}

	return nil
}

var testAccCloudStackSSHKey_create = fmt.Sprintf(`
resource "cloudstack_ssh_key" "foo" {
  name = "terraform-testacc"
}`)
