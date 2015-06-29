package netapi

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform/helper/schema"
	"github.com/hashicorp/terraform/terraform"
)

var testAccProviders map[string]terraform.ResourceProvider
var testAccProvider *schema.Provider

func init() {
	testAccProvider = Provider().(*schema.Provider)
	testAccProviders = map[string]terraform.ResourceProvider{
		"netapi": testAccProvider,
	}
}

func TestProvider(t *testing.T) {
	if err := Provider().(*schema.Provider).InternalValidate(); err != nil {
		t.Fatalf("err: %s", err)
	}
}

func TestProvider_impl(t *testing.T) {
	var _ terraform.ResourceProvider = Provider()
}

func testAccPreCheck(t *testing.T) {
	if v := os.Getenv("NETAPI_URL"); v == "" {
		t.Fatal("NETAPI_URL must be set for acceptance tests")
	}
	if v := os.Getenv("NETAPI_PUBLIC_KEY"); v == "" {
		t.Fatal("NETAPI_PUBLIC_KEY must be set for acceptance tests")
	}
	if v := os.Getenv("NETAPI_SECRET_KEY"); v == "" {
		t.Fatal("NETAPI_SECRET_KEY must be set for acceptance tests")
	}
	if v := os.Getenv("NETAPI_SECRET_KEY"); v == "" {
		t.Fatal("NETAPI_SECRET_KEY must be set for acceptance tests")
	}
	if v := os.Getenv("NETAPI_ZONE"); v == "" {
		t.Fatal("NETAPI_ZONE must be set for acceptance tests")
	}
	if v := os.Getenv("NETAPI_DCG"); v == "" {
		t.Fatal("NETAPI_DCG must be set for acceptance tests")
	}

}

var NETAPI_ZONE = os.Getenv("NETAPI_ZONE")
var NETAPI_DCG = os.Getenv("NETAPI_DCG")
