package netapi

import (
	"fmt"
	"testing"

	"github.com/benjvi/go-net-api"
	"github.com/hashicorp/terraform/helper/resource"
	"github.com/hashicorp/terraform/terraform"
)

func TestAccNetAPIPrivateDirectConnect_createAndUpdate(t *testing.T) {
	var vpn netAPI.Network

	resource.Test(t, resource.TestCase{
		PreCheck:  func() { testAccPreCheck(t) },
		Providers: testAccProviders,
		Steps: []resource.TestStep{
			resource.TestStep{
				Config: testAccNetAPIPrivateDirectConnect_one,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNetAPIPrivateDirectConnectExists(
						"netapi_private_direct_connect.foo", &vpn),
					testAccCheckNetAPIPrivateDirectConnectAttributesOne(&vpn),
				),
			},
			resource.TestStep{
				Config: testAccNetAPIPrivateDirectConnect_two,
				Check: resource.ComposeTestCheckFunc(
					testAccCheckNetAPIPrivateDirectConnectExists(
						"netapi_private_direct_connect.foo", &vpn),
					testAccCheckNetAPIPrivateDirectConnectAttributesTwo(&vpn),
				),
			},
		},
	})
}

func testAccCheckNetAPIPrivateDirectConnectExists(
	n string, vpn *netAPI.Network) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		rs, ok := s.RootModule().Resources[n]
		if !ok {
			return fmt.Errorf("Not found: %s", n)
		}

		if rs.Primary.ID == "" {
			return fmt.Errorf("No VPN ID is set")
		}

		cs := testAccProvider.Meta().(*netAPI.NetAPIClient)
		network, _, err := cs.Network.GetNetworkByID(rs.Primary.ID, "europe")

		if err != nil {
			return err
		}

		if network.Id != rs.Primary.ID {
			return fmt.Errorf("VPN not found")
		}

		*vpn = *network

		return nil
	}
}

func testAccCheckNetAPIPrivateDirectConnectAttributesOne(
	vpn *netAPI.Network) resource.TestCheckFunc {
	return func(s *terraform.State) error {

		if vpn.Displaytext != "terraform-acctest-vpn-1" {
			return fmt.Errorf("Bad name: %s", vpn.Displaytext)
		}

		if vpn.Cidr != "10.0.31.0/24" {
			return fmt.Errorf("Bad CIDR: %s", vpn.Cidr)
		}

		if vpn.Gateway != "10.0.31.1" {
			return fmt.Errorf("Bad gateway: %s", vpn.Gateway)
		}

		if vpn.Zonename != NETAPI_ZONE {
			return fmt.Errorf("Bad zone: %s", vpn.Zonename)
		}

		if vpn.Dcgfriendlyname != NETAPI_DCG {
			return fmt.Errorf("Bad DCG: %s", vpn.Dcgfriendlyname)
		}

		return nil
	}
}

func testAccCheckNetAPIPrivateDirectConnectAttributesTwo(
	vpn *netAPI.Network) resource.TestCheckFunc {
	return func(s *terraform.State) error {

		if vpn.Displaytext != "terraform-acctest-vpn-2" {
			return fmt.Errorf("Bad name: %s", vpn.Displaytext)
		}

		if vpn.Cidr != "10.0.32.0/24" {
			return fmt.Errorf("Bad CIDR: %s", vpn.Cidr)
		}

		if vpn.Gateway != "10.0.32.1" {
			return fmt.Errorf("Bad gateway: %s", vpn.Gateway)
		}

		if vpn.Zonename != NETAPI_ZONE {
			return fmt.Errorf("Bad zone: %s", vpn.Zonename)
		}

		if vpn.Dcgfriendlyname != NETAPI_DCG {
			return fmt.Errorf("Bad DCG: %s", vpn.Dcgfriendlyname)
		}

		return nil
	}
}

var testAccNetAPIPrivateDirectConnect_one = fmt.Sprintf(`
resource "netapi_private_direct_connect" "foo" {
    display_text = "terraform-acctest-vpn-1"
    zone = "%s"
    cidr = "10.0.31.0/24"
    gateway = "10.0.31.1"
    dcg = "%s"
    region = 'europe'
}`,
	NETAPI_ZONE,
	NETAPI_DCG)

var testAccNetAPIPrivateDirectConnect_two = fmt.Sprintf(`
resource "netapi_private_direct_connect" "foo" {
    display_text = "terraform-acctest-vpn-2"
    zone = "%s"
    cidr = "10.0.32.0/24"
    gateway = "10.0.32.1"
    dcg = "%s"
    region = "europe"
}`,
	NETAPI_ZONE,
	NETAPI_DCG)
