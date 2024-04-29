locals {
  org_id    = "example_org_id"
  bucket_id = "example_bucket_id"
}

resource "influxdb-v2_dbrp_mapping" "example_dbrp_mapping" {
  org_id           = local.org_id
  bucket_id        = local.bucket_id
  database         = "legacy_database_name"
  retention_policy = "legacy_retention_policy_name"
}
