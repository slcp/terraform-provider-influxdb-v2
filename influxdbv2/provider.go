package influxdbv2

import (
	"context"
	"crypto/tls"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
)

type meta struct {
	influxsdk                  influxdb2.Client
	legacyAuthorizationsClient *Client
}

func Provider() *schema.Provider {
	return &schema.Provider{
		DataSourcesMap: map[string]*schema.Resource{
			"influxdb-v2_ready":        DataReady(),
			"influxdb-v2_organization": dataSourceOrganization(),
			"influxdb-v2_bucket":       dataSourceBucket(),
		},
		ResourcesMap: map[string]*schema.Resource{
			"influxdb-v2_bucket":        ResourceBucket(),
			"influxdb-v2_authorization": ResourceAuthorization(),
			"influxdb-v2_organization":  ResourceOrganization(),
		},
		Schema: map[string]*schema.Schema{
			"url": {
				Type:     schema.TypeString,
				Optional: true,
				DefaultFunc: schema.EnvDefaultFunc(
					"INFLUXDB_V2_URL",
					"http://localhost:8086",
				),
			},
			"token": {
				Type:        schema.TypeString,
				Optional:    true,
				Sensitive:   true,
				DefaultFunc: schema.EnvDefaultFunc("INFLUXDB_V2_TOKEN", ""),
			},
			"skip_ssl_verify": {
				Type:        schema.TypeBool,
				Description: "skip ssl verify on connection",
				Optional:    true,
				DefaultFunc: schema.EnvDefaultFunc("INFLUXDB_SKIP_SSL_VERIFY", "0"),
			},
		},
		ConfigureFunc: providerConfigure,
	}
}

func providerConfigure(d *schema.ResourceData) (interface{}, error) {
	url := d.Get("url").(string)
	token := d.Get("token").(string)
	sslv := d.Get("skip_ssl_verify").(bool)
	tls := &tls.Config{
		InsecureSkipVerify: sslv,
	}
	opts := influxdb2.DefaultOptions().SetTLSConfig(tls)
	influx := influxdb2.NewClientWithOptions(url, token, opts)

	_, err := influx.Ready(context.Background())
	if err != nil {
		return nil, fmt.Errorf("error pinging server: %s", err)
	}

	legacy, err := NewClient(url)
	if err != nil {
		return nil, fmt.Errorf("error creating legacy client: %s", err)
	}
	return meta{
		influxsdk:                  influx,
		legacyAuthorizationsClient: legacy,
	}, nil
}
