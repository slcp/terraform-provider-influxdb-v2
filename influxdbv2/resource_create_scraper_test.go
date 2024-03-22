package influxdbv2

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/domain"
)

var scraperIdOnCreate string

func TestAccCreateScraper(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccScraperDestroyed,
		Steps: []resource.TestStep{
			{
				Config: testAccCreateScraper(),
				Check: resource.ComposeTestCheckFunc(
					func(s *terraform.State) error {
						id := extractIdForResource(s, "influxdb-v2_scraper.acctest")
						scraperIdOnCreate = id
						return nil
					},
					resource.TestCheckResourceAttr(
						"influxdb-v2_scraper.acctest",
						"org_id",
						os.Getenv("INFLUXDB_V2_ORG_ID"),
					),
					resource.TestCheckResourceAttr(
						"influxdb-v2_scraper.acctest",
						"name",
						"acctest",
					),
					resource.TestCheckResourceAttr(
						"influxdb-v2_scraper.acctest",
						"allow_insecure",
						"true",
					),
					resource.TestCheckResourceAttr(
						"influxdb-v2_scraper.acctest",
						"url",
						"http://localhost:8086/metrics",
					),
					resource.TestCheckResourceAttr(
						"influxdb-v2_scraper.acctest",
						"type",
						"prometheus",
					),
				),
			},
			{
				// This test is designed to prove that the provider no longer errors when asked to read a resource that doesn't exist.
				// The desired approach is to signal to terraform that the resource cannot be found so that the plan is to recreate it.
				Config: testAccCreateScraper(),
				PreConfig: func() {
					deleteScraper("acctest")
				},
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttrSet("influxdb-v2_scraper.acctest", "id"),
					checkResourceHasBeenReplaced("influxdb-v2_scraper.acctest", &scraperIdOnCreate),
				),
			},
			{
				Config: testAccUpdateScraper(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(
						"influxdb-v2_scraper.acctest",
						"org_id",
						os.Getenv("INFLUXDB_V2_ORG_ID"),
					),
					resource.TestCheckResourceAttr(
						"influxdb-v2_scraper.acctest",
						"name",
						"acctest2",
					),
					resource.TestCheckResourceAttr(
						"influxdb-v2_scraper.acctest",
						"allow_insecure",
						"false",
					),
					resource.TestCheckResourceAttr(
						"influxdb-v2_scraper.acctest",
						"url",
						"http://localhost:8086/metrics2",
					),
					testAccCheckUpdate("influxdb-v2_scraper.acctest"),
					resource.TestCheckResourceAttr(
						"influxdb-v2_scraper.acctest",
						"type",
						"prometheus",
					),
				),
			},
		},
	})
}

func testAccCreateScraper() string {
	return `
resource "influxdb-v2_bucket" "acctest" {
	description = "Acceptance test bucket" 
	name = "acctest" 
	org_id = "` + os.Getenv("INFLUXDB_V2_ORG_ID") + `"
	retention_rules {
		every_seconds = "3640"
	}
}
resource "influxdb-v2_scraper" "acctest" {
    name = "acctest" 
    org_id = "` + os.Getenv("INFLUXDB_V2_ORG_ID") + `"
    allow_insecure = true
	bucket_id = influxdb-v2_bucket.acctest.id
	url = "http://localhost:8086/metrics"
}
`
}

func testAccUpdateScraper() string {
	return `
resource "influxdb-v2_bucket" "acctest" {
	description = "Acceptance test bucket" 
	name = "acctest" 
	org_id = "` + os.Getenv("INFLUXDB_V2_ORG_ID") + `"
	retention_rules {
		every_seconds = "3640"
	}
}
resource "influxdb-v2_scraper" "acctest" {
	name = "acctest2" 
	org_id = "` + os.Getenv("INFLUXDB_V2_ORG_ID") + `"
	allow_insecure = false
	bucket_id = influxdb-v2_bucket.acctest.id
	url = "http://localhost:8086/metrics2"
}
`
}

func testAccScraperDestroyed(s *terraform.State) error {
	influx := influxdb2.NewClient(
		os.Getenv("INFLUXDB_V2_URL"),
		os.Getenv("INFLUXDB_V2_TOKEN"),
	)
	result, err := influx.BucketsAPI().GetBuckets(context.Background())
	// The only buckets are from the onboarding, plus _monitoring and _tasks
	if len(*result) != 3 {
		return fmt.Errorf(
			"There should be only one remaining bucket but there are: %d",
			len(*result),
		)
	}
	if err != nil {
		return fmt.Errorf("Cannot read bucket list")
	}
	return nil
}

func deleteScraper(name string) {
	influx := influxdb2.NewClient(
		os.Getenv("INFLUXDB_V2_URL"),
		os.Getenv("INFLUXDB_V2_TOKEN"),
	)
	result, err := influx.APIClient().GetScrapers(context.Background(), &domain.GetScrapersParams{
		Name: &name,
	})
	if err != nil {
		panic("Cannot find scraper")
	}

	if len(*result.Configurations) > 1 {
		panic("There should be only one scraper")
	}

	scraper := (*result.Configurations)[0]

	err = influx.APIClient().DeleteScrapersID(context.Background(), &domain.DeleteScrapersIDAllParams{
		ScraperTargetID: *scraper.Id,
	})
	if err != nil {
		panic("Cannot delete scraper")
	}
}
