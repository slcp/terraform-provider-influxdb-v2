provider "influxdb-v2" {
  url             = "http://localhost:8086"    # changeme
  token           = "super-secret-admin-token" # changeme
  skip_ssl_verify = true                       # optional
  health_check    = "ping"                     # optional
}
