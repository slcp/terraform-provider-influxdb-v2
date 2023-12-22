locals {
  org_id    = "example_org_id"
  bucket_id = "example_bucket_id"
}

resource "influxdb-v2_dbrp_mapping" "example_dbrp_mapping" {
	org_id = var.org_id
	bucket_id = var.bucket_id
	database = "legacy_database"
	retention_policy = "6w"
}