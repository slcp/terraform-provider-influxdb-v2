package influxdbv2

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

var idOnCreate string

func TestAccLegacyAuthorization(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccLegacyAuthorizationDestroyed,
		Steps: []resource.TestStep{
			{
				Config: testAccCreateLegacyAuthorization(),
				Check: resource.ComposeTestCheckFunc(
					func(s *terraform.State) error {
						id := extractIdForResource(s, "influxdb-v2_legacy_authorization.acctest")
						idOnCreate = id
						return nil
					},
					resource.TestCheckResourceAttr("influxdb-v2_legacy_authorization.acctest", "org_id", os.Getenv("INFLUXDB_V2_ORG_ID")),
					resource.TestCheckResourceAttr("influxdb-v2_legacy_authorization.acctest", "description", "Acceptance test legacy token"),
					resource.TestCheckResourceAttr("influxdb-v2_legacy_authorization.acctest", "status", "inactive"),
					resource.TestCheckResourceAttrSet("influxdb-v2_legacy_authorization.acctest", "user_id"),
					resource.TestCheckResourceAttr("influxdb-v2_legacy_authorization.acctest", "user_org_id", os.Getenv("INFLUXDB_V2_ORG_ID")),
					resource.TestCheckResourceAttrSet("influxdb-v2_legacy_authorization.acctest", "name"),
					resource.TestCheckResourceAttr("influxdb-v2_legacy_authorization.acctest", "org_id", os.Getenv("INFLUXDB_V2_ORG_ID")),
					// permissions is a complex array... we'll just check it has the correct length
					resource.TestCheckResourceAttr("influxdb-v2_legacy_authorization.acctest", "permissions.#", "2"),
				),
			},
			{
				Config: testAccLegacyUpdateAuthorization(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("influxdb-v2_legacy_authorization.acctest", "org_id", os.Getenv("INFLUXDB_V2_ORG_ID")),
					resource.TestCheckResourceAttr("influxdb-v2_legacy_authorization.acctest", "description", "Acceptance test legacy token 2"),
					resource.TestCheckResourceAttr("influxdb-v2_legacy_authorization.acctest", "status", "active"),
					resource.TestCheckResourceAttrSet("influxdb-v2_legacy_authorization.acctest", "user_id"),
					resource.TestCheckResourceAttr("influxdb-v2_legacy_authorization.acctest", "user_org_id", os.Getenv("INFLUXDB_V2_ORG_ID")),
					resource.TestCheckResourceAttrSet("influxdb-v2_legacy_authorization.acctest", "name"),
					resource.TestCheckResourceAttr("influxdb-v2_legacy_authorization.acctest", "org_id", os.Getenv("INFLUXDB_V2_ORG_ID")),
					// permissions is a complex array... we'll just check it has the correct length
					resource.TestCheckResourceAttr("influxdb-v2_legacy_authorization.acctest", "permissions.#", "2"),
				),
			},
			{
				Config: testAccLegacyUpdateAuthorizationPermissions(),
				Check: resource.ComposeTestCheckFunc(
					checkResourceHasBeenReplaced("influxdb-v2_legacy_authorization.acctest", &idOnCreate),
				),
			},
		},
	})
}

func testAccCreateLegacyAuthorization() string {
	return `
resource "influxdb-v2_legacy_authorization" "acctest" {
	org_id = "` + os.Getenv("INFLUXDB_V2_ORG_ID") + `"
    description = "Acceptance test legacy token"
	name = "a user name"
	password = "secure password"
    status = "inactive"
    permissions {
        action = "read"
        resource {
			id = "` + os.Getenv("INFLUXDB_V2_BUCKET_ID") + `"
            org_id = "` + os.Getenv("INFLUXDB_V2_ORG_ID") + `"
            type = "buckets"
        }
    }
    permissions {
        action = "write"
        resource {
            id = "` + os.Getenv("INFLUXDB_V2_BUCKET_ID") + `"
            org_id = "` + os.Getenv("INFLUXDB_V2_ORG_ID") + `"
            type = "buckets"
        }
    }
}
`
}

func testAccLegacyUpdateAuthorization() string {
	return `
resource "influxdb-v2_legacy_authorization" "acctest" {
	org_id = "` + os.Getenv("INFLUXDB_V2_ORG_ID") + `"
	description = "Acceptance test legacy token 2"
	name = "a user name"
	password = "another secure password"
	status = "active"
	permissions {
		action = "read"
		resource {
			id = "` + os.Getenv("INFLUXDB_V2_BUCKET_ID") + `"
			org_id = "` + os.Getenv("INFLUXDB_V2_ORG_ID") + `"
			type = "buckets"
		}
	}
	permissions {
		action = "write"
		resource {
			id = "` + os.Getenv("INFLUXDB_V2_BUCKET_ID") + `"
			org_id = "` + os.Getenv("INFLUXDB_V2_ORG_ID") + `"
			type = "buckets"
		}
	}
}
`
}

func testAccLegacyUpdateAuthorizationPermissions() string {
	return `
resource "influxdb-v2_legacy_authorization" "acctest" {
	org_id = "` + os.Getenv("INFLUXDB_V2_ORG_ID") + `"
	description = "Acceptance test legacy token 2"
	name = "a user name"
	password = "another secure password"
	status = "active"
	permissions {
		action = "read"
		resource {
			id = "` + os.Getenv("INFLUXDB_V2_BUCKET_ID") + `"
			org_id = "` + os.Getenv("INFLUXDB_V2_ORG_ID") + `"
			type = "buckets"
		}
	}
}
`
}

func testAccLegacyAuthorizationDestroyed(s *terraform.State) error {
	addToken := func(ctx context.Context, req *http.Request) error {
		req.Header.Add("Authorization", fmt.Sprintf("Token %s", os.Getenv("INFLUXDB_V2_TOKEN")))
		return nil
	}
	influx, err := NewClientWithResponses(fmt.Sprint(os.Getenv("INFLUXDB_V2_URL"), "/private"), WithRequestEditorFn(addToken))
	result, err := influx.GetLegacyAuthorizationsWithResponse(context.Background(), &GetLegacyAuthorizationsParams{})

	if len(*result.JSON200.Authorizations) != 0 {
		return fmt.Errorf("There should be only one remaining authorization but there are: %d", len(*result.JSON200.Authorizations))
	}
	if err != nil {
		return fmt.Errorf("Cannot read authorization list")
	}

	return nil
}

func extractIdForResource(s *terraform.State, name string) string {
	is := findResourceInState(s, name)
	v, _ := is.Attributes["id"]
	return v
}

func findResourceInState(s *terraform.State, name string) *terraform.InstanceState {
	ms := s.RootModule()
	rs, ok := ms.Resources[name]
	if !ok {
		return nil
	}
	is := rs.Primary
	return is
}

func checkResourceHasBeenReplaced(name string, oid *string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if id := extractIdForResource(s, name); id == *oid {
			return fmt.Errorf("idd should have changed but it is still %s", id)
		}
		return nil
	}
}
