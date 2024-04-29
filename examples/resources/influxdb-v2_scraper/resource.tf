locals {
  org_id    = "example_org_id"
  bucket_id = "example_bucket_id"
}

resource "influxdb-v2_scraper" "example_scraper" {
  name           = "example_scraper"
  org_id         = local.org_id
  allow_insecure = false
  bucket_id      = local.bucket_id
  url            = "http://scraper-target.com/metrics"
}
