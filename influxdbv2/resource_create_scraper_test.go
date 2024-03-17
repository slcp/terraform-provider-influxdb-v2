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

// var bucketIdOnCreate string

func TestAccCreateScraper(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:     func() { testAccPreCheck(t) },
		Providers:    testAccProviders,
		CheckDestroy: testAccScraperDestroyed,
		Steps: []resource.TestStep{
			{
				Config: testAccCreateScraper(),
				Check: resource.ComposeTestCheckFunc(
					// func(s *terraform.State) error {
					// 	id := extractIdForResource(s, "influxdb-v2_bucket.acctest")
					// 	bucketIdOnCreate = id
					// 	return nil
					// },
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
					testAccCheckUpdate("influxdb-v2_scraper.acctest"),
					resource.TestCheckResourceAttr(
						"influxdb-v2_scraper.acctest",
						"type",
						"prometheus",
					),
				),
			},
			// {
			// 	// This test is designed to prove that the provider no longer errors when asked to read a resource that doesn't exist.
			// 	// The desired approach is to signal to terraform that the resource cannot be found so that the plan is to recreate it.
			// 	Config: testAccCreateScraper(),
			// 	PreConfig: func() {
			// 		deleteScraper("acctest")
			// 	},
			// 	Check: resource.ComposeTestCheckFunc(
			// 		resource.TestCheckResourceAttrSet("influxdb-v2_scraper.acctest", "id"),
			// 		checkResourceHasBeenReplaced("influxdb-v2_scraper.acctest", &bucketIdOnCreate),
			// 	),
			// },
			// {
			// 	Config: testAccUpdateScraper(),
			// 	Check: resource.ComposeTestCheckFunc(
			// 		resource.TestCheckResourceAttr(
			// 			"influxdb-v2_scraper.acctest",
			// 			"org_id",
			// 			os.Getenv("INFLUXDB_V2_ORG_ID"),
			// 		),
			// 		resource.TestCheckResourceAttr(
			// 			"influxdb-v2_scraper.acctest",
			// 			"description",
			// 			"Acceptance test bucket 2",
			// 		),
			// 		resource.TestCheckResourceAttr(
			// 			"influxdb-v2_scraper.acctest",
			// 			"name",
			// 			"acctest",
			// 		),
			// 		resource.TestCheckResourceAttr(
			// 			"influxdb-v2_scraper.acctest",
			// 			"rp",
			// 			"",
			// 		),
			// 		resource.TestCheckResourceAttr(
			// 			"influxdb-v2_scraper.acctest",
			// 			"retention_rules.0.every_seconds",
			// 			"3630",
			// 		),
			// 		resource.TestCheckResourceAttrSet(
			// 			"influxdb-v2_scraper.acctest",
			// 			"created_at",
			// 		),
			// 		resource.TestCheckResourceAttrSet(
			// 			"influxdb-v2_scraper.acctest",
			// 			"updated_at",
			// 		),
			// 		testAccCheckUpdate("influxdb-v2_scraper.acctest"),
			// 		resource.TestCheckResourceAttrSet(
			// 			"influxdb-v2_scraper.acctest",
			// 			"type",
			// 		),
			// 	),
			// },
		},
	})
}

// var lastUpdate = ""

//	func testAccCheckUpdate(n string) resource.TestCheckFunc {
//		return func(s *terraform.State) error {
//			rs, ok := s.RootModule().Resources[n]
//			if !ok {
//				return fmt.Errorf("Resource %s doesn't exist", n)
//			}
//			if lastUpdate == rs.Primary.Attributes["updated_at"] {
//				return fmt.Errorf("updated_at has not changed since last execution")
//			}
//			lastUpdate = rs.Primary.Attributes["updated_at"]
//			return nil
//		}
//	}
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

// func testAccUpdateScraper() string {
// 	return `
// resource "influxdb-v2_scraper" "acctest" {
//     description = "Acceptance test bucket 2"
//     name = "acctest"
//     org_id = "` + os.Getenv("INFLUXDB_V2_ORG_ID") + `"
//     retention_rules {
//         every_seconds = "3630"
//     }
// }
// `
// }

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

// func deleteScraper(name string) {
// 	influx := influxdb2.NewClient(
// 		os.Getenv("INFLUXDB_V2_URL"),
// 		os.Getenv("INFLUXDB_V2_TOKEN"),
// 	)
// 	result, err := influx.BucketsAPI().FindBucketByName(context.Background(), name)
// 	if err != nil {
// 		panic("Cannot find bucket")
// 	}

// 	err = influx.BucketsAPI().DeleteScraper(context.Background(), result)
// 	if err != nil {
// 		panic("Cannot delete bucket")
// 	}
// }
