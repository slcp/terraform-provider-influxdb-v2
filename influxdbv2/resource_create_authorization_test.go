package influxdbv2

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
)

var authorizationIdOnCreate string

func TestAccAuthorization(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccAuthorizationDestroyed,
		Steps: []resource.TestStep{
			{
				Config: testAccCreateAuthorization(),
				Check: resource.ComposeTestCheckFunc(
					func(s *terraform.State) error {
						id := extractIdForResource(s, "influxdb-v2_authorization.acctest")
						authorizationIdOnCreate = id
						return nil
					},
					resource.TestCheckResourceAttr("influxdb-v2_authorization.acctest", "org_id", os.Getenv("INFLUXDB_V2_ORG_ID")),
					resource.TestCheckResourceAttr("influxdb-v2_authorization.acctest", "description", "Acceptance test token"),
					resource.TestCheckResourceAttr("influxdb-v2_authorization.acctest", "status", "inactive"),
					resource.TestCheckResourceAttrSet("influxdb-v2_authorization.acctest", "user_id"),
					resource.TestCheckResourceAttr("influxdb-v2_authorization.acctest", "user_org_id", os.Getenv("INFLUXDB_V2_ORG_ID")),
					resource.TestCheckResourceAttrSet("influxdb-v2_authorization.acctest", "token"),
					resource.TestCheckResourceAttr("influxdb-v2_authorization.acctest", "org_id", os.Getenv("INFLUXDB_V2_ORG_ID")),
					// permissions is a complex array... we'll just check it has the correct length
					resource.TestCheckResourceAttr("influxdb-v2_authorization.acctest", "permissions.#", "2"),
				),
			},
			{
				Config: testAccUpdateAuthorization(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("influxdb-v2_authorization.acctest", "org_id", os.Getenv("INFLUXDB_V2_ORG_ID")),
					resource.TestCheckResourceAttr("influxdb-v2_authorization.acctest", "description", "Acceptance test token 2"),
					resource.TestCheckResourceAttr("influxdb-v2_authorization.acctest", "status", "active"),
					resource.TestCheckResourceAttrSet("influxdb-v2_authorization.acctest", "user_id"),
					resource.TestCheckResourceAttr("influxdb-v2_authorization.acctest", "user_org_id", os.Getenv("INFLUXDB_V2_ORG_ID")),
					resource.TestCheckResourceAttrSet("influxdb-v2_authorization.acctest", "token"),
					resource.TestCheckResourceAttr("influxdb-v2_authorization.acctest", "org_id", os.Getenv("INFLUXDB_V2_ORG_ID")),
					// permissions is a complex array... we'll just check it has the correct length
					resource.TestCheckResourceAttr("influxdb-v2_authorization.acctest", "permissions.#", "2"),
				),
			},
			{
				// This test is designed to prove that the provider no longer errors when asked to read a resource that doesn't exist.
				// The desired approach is to signal to terraform that the resource cannot be found so that the plan is to recreate it.
				Config: testAccUpdateAuthorization(),
				PreConfig: func() {
					deleteAuth(authorizationIdOnCreate)
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("influxdb-v2_authorization.acctest", "id"),
					checkResourceHasBeenReplaced("influxdb-v2_authorization.acctest", &authorizationIdOnCreate),
				),
			},
		},
	})
}

func testAccCreateAuthorization() string {
	return `
resource "influxdb-v2_authorization" "acctest" {
	org_id = "` + os.Getenv("INFLUXDB_V2_ORG_ID") + `"
    description = "Acceptance test token"
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

func testAccUpdateAuthorization() string {
	return `
resource "influxdb-v2_authorization" "acctest" {
	org_id = "` + os.Getenv("INFLUXDB_V2_ORG_ID") + `"
    description = "Acceptance test token 2"
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

func testAccAuthorizationDestroyed(s *terraform.State) error {
	influx := influxdb2.NewClient(os.Getenv("INFLUXDB_V2_URL"), os.Getenv("INFLUXDB_V2_TOKEN"))
	result, err := influx.AuthorizationsAPI().GetAuthorizations(context.Background())
	// The only auth is from the onboarding
	if len(*result) != 1 {
		return fmt.Errorf("There should be only one remaining authorization but there are: %d", len(*result))
	}
	if err != nil {
		return fmt.Errorf("Cannot read authorization list")
	}

	return nil
}

func deleteAuth(id string) {
	influx := influxdb2.NewClient(
		os.Getenv("INFLUXDB_V2_URL"),
		os.Getenv("INFLUXDB_V2_TOKEN"),
	)
	err := influx.AuthorizationsAPI().DeleteAuthorizationWithID(context.Background(), id)
	if err != nil {
		panic(fmt.Sprintf("Cannot delete authorization: %v", err))
	}
}
