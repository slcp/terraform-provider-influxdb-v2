package influxdbv2

import (
	"context"
	"crypto/tls"
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/domain"
)

var dbrpIdOnCreate string

func TestAccDBRPMapping(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccDBRPMappingDestroyed,
		Steps: []resource.TestStep{
			{
				Config: testAccCreateDBRPMapping(),
				Check: resource.ComposeTestCheckFunc(
					func(s *terraform.State) error {
						id := extractIdForResource(s, "influxdb-v2_dbrp_mapping.acctest")
						dbrpIdOnCreate = id
						return nil
					},
					resource.TestCheckResourceAttr("influxdb-v2_dbrp_mapping.acctest", "org_id", os.Getenv("INFLUXDB_V2_ORG_ID")),
					resource.TestCheckResourceAttr("influxdb-v2_dbrp_mapping.acctest", "bucket_id", os.Getenv("INFLUXDB_V2_BUCKET_ID")),
					resource.TestCheckResourceAttr("influxdb-v2_dbrp_mapping.acctest", "database", "legacy_database"),
					resource.TestCheckResourceAttrSet("influxdb-v2_dbrp_mapping.acctest", "id"),
					resource.TestCheckResourceAttr("influxdb-v2_dbrp_mapping.acctest", "retention_policy", "legacy_rp"),
					resource.TestCheckResourceAttr("influxdb-v2_dbrp_mapping.acctest", "default_policy", "true"),
				),
			},
			{
				// This test is designed to prove that the provider no longer errors when asked to read a resource that doesn't exist.
				// The desired approach is to signal to terraform that the resource cannot be found so that the plan is to recreate it.
				Config: testAccCreateDBRPMapping(),
				PreConfig: func() {
					deleteDBRP(dbrpIdOnCreate)
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("influxdb-v2_dbrp_mapping.acctest", "id"),
					checkResourceHasBeenReplaced("influxdb-v2_dbrp_mapping.acctest", &dbrpIdOnCreate),
				),
			},
			{
				Config: testAccUpdateDBRPMapping(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("influxdb-v2_dbrp_mapping.acctest", "org_id", os.Getenv("INFLUXDB_V2_ORG_ID")),
					resource.TestCheckResourceAttr("influxdb-v2_dbrp_mapping.acctest", "bucket_id", os.Getenv("INFLUXDB_V2_BUCKET_ID")),
					resource.TestCheckResourceAttr("influxdb-v2_dbrp_mapping.acctest", "database", "legacy_database"),
					resource.TestCheckResourceAttrSet("influxdb-v2_dbrp_mapping.acctest", "id"),
					resource.TestCheckResourceAttr("influxdb-v2_dbrp_mapping.acctest", "retention_policy", "some_other_legacy_rp"),
					resource.TestCheckResourceAttr("influxdb-v2_dbrp_mapping.acctest", "default_policy", "true"),
				),
			},
		},
	})
}

func testAccCreateDBRPMapping() string {
	return `
resource "influxdb-v2_dbrp_mapping" "acctest" {
	org_id = "` + os.Getenv("INFLUXDB_V2_ORG_ID") + `"
	bucket_id = "` + os.Getenv("INFLUXDB_V2_BUCKET_ID") + `"
	database = "legacy_database"
	retention_policy = "legacy_rp"
}
`
}

func testAccUpdateDBRPMapping() string {
	return `
resource "influxdb-v2_dbrp_mapping" "acctest" {
	org_id = "` + os.Getenv("INFLUXDB_V2_ORG_ID") + `"
	bucket_id = "` + os.Getenv("INFLUXDB_V2_BUCKET_ID") + `"
	database = "legacy_database"
	retention_policy = "some_other_legacy_rp"
}
`
}

func testAccDBRPMappingDestroyed(s *terraform.State) error {
	orgId := os.Getenv("INFLUXDB_V2_ORG_ID")
	tls := &tls.Config{
		InsecureSkipVerify: true,
	}
	opts := influxdb2.DefaultOptions().SetTLSConfig(tls)
	influx := influxdb2.NewClientWithOptions(os.Getenv("INFLUXDB_V2_URL"), os.Getenv("INFLUXDB_V2_TOKEN"), opts)
	result, err := influx.APIClient().GetDBRPs(context.Background(), &domain.GetDBRPsParams{
		OrgID: &orgId,
	})
	if err != nil {
		return fmt.Errorf("error getting dbrps: %v", err)
	}
	for _, dbrp := range *result.Content {
		if dbrp.BucketID == os.Getenv("INFLUXDB_V2_BUCKET_ID") && dbrp.RetentionPolicy != "autogen" {
			return fmt.Errorf("There should only be autogen dbrp mapping for bucket but there is one called: %s", dbrp.RetentionPolicy)
		}
	}

	return nil
}

func deleteDBRP(id string) {
	orgId := os.Getenv("INFLUXDB_V2_ORG_ID")
	influx := influxdb2.NewClient(
		os.Getenv("INFLUXDB_V2_URL"),
		os.Getenv("INFLUXDB_V2_TOKEN"),
	)
	err := influx.APIClient().DeleteDBRPID(context.Background(), &domain.DeleteDBRPIDAllParams{
		DbrpID: id,
		DeleteDBRPIDParams: domain.DeleteDBRPIDParams{
			OrgID: &orgId,
		},
	})
	if err != nil {
		panic("Cannot delete dbrp")
	}
}
