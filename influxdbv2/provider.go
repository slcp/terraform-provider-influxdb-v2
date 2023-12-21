package influxdbv2

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
)

type meta struct {
	influxsdk                  influxdb2.Client
	legacyAuthorizationsClient *ClientWithResponses
}

func Provider() *schema.Provider {
	return &schema.Provider{
		DataSourcesMap: map[string]*schema.Resource{
			"influxdb-v2_ready":        DataReady(),
			"influxdb-v2_organization": dataSourceOrganization(),
			"influxdb-v2_bucket":       dataSourceBucket(),
		},
		ResourcesMap: map[string]*schema.Resource{
			"influxdb-v2_bucket":               ResourceBucket(),
			"influxdb-v2_authorization":        ResourceAuthorization(),
			"influxdb-v2_organization":         ResourceOrganization(),
			"influxdb-v2_legacy_authorization": ResourceLegacyAuthorization(),
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

	addToken := func(ctx context.Context, req *http.Request) error {
		req.Header.Add("Authorization", fmt.Sprintf("Token %s", token))
		return nil
	}
	skipSSLVerify := func(client *Client) error {
		httpClient := &http.Client{}
		httpClient.Transport = &http.Transport{
			TLSClientConfig: tls,
		}
		client.Client = httpClient
		return nil
	}
	legacy, err := NewClientWithResponses(fmt.Sprint(url, "/private"), WithRequestEditorFn(addToken), skipSSLVerify)
	if err != nil {
		return nil, fmt.Errorf("error creating legacy client: %s", err)
	}
	return meta{
		influxsdk:                  influx,
		legacyAuthorizationsClient: legacy,
	}, nil
}
