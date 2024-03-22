package influxdbv2

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/influxdata/influxdb-client-go/v2/domain"
)

func ResourceScraper() *schema.Resource {
	return &schema.Resource{
		Create: resourceScraperCreate,
		Delete: resourceScraperDelete,
		Read:   resourceScraperRead,
		Update: resourceScraperUpdate,
		Schema: map[string]*schema.Schema{
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"org_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"bucket_id": {
				Type:     schema.TypeString,
				Required: true,
			},
			"url": {
				Type:     schema.TypeString,
				Required: true,
			},
			"type": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"allow_insecure": {
				Type:     schema.TypeBool,
				Required: true,
			},
		},
	}
}

func resourceScraperCreate(d *schema.ResourceData, m interface{}) error {
	influx := m.(meta).influxsdk
	orgid := d.Get("org_id").(string)
	bucketid := d.Get("bucket_id").(string)
	name := d.Get("name").(string)
	insecure := d.Get("allow_insecure").(bool)
	targettype := domain.ScraperTargetRequestType("prometheus")
	url := d.Get("url").(string)

	newScraper := &domain.PostScrapersAllParams{
		Body: domain.PostScrapersJSONRequestBody{
			OrgID:         &orgid,
			BucketID:      &bucketid,
			Name:          &name,
			AllowInsecure: &insecure,
			Type:          &targettype,
			Url:           &url,
		},
	}
	result, err := influx.APIClient().PostScrapers(context.Background(), newScraper)
	if err != nil {
		return fmt.Errorf("error creating Scraper: %v", err)
	}
	d.SetId(*result.Id)
	return resourceScraperRead(d, m)
}

func resourceScraperDelete(d *schema.ResourceData, m interface{}) error {
	influx := m.(meta).influxsdk
	err := influx.APIClient().DeleteScrapersID(context.Background(), &domain.DeleteScrapersIDAllParams{
		ScraperTargetID: d.Id(),
	})
	if err != nil {
		return fmt.Errorf("error deleting Scraper: %v", err)
	}
	d.SetId("")
	return nil
}

func resourceScraperRead(d *schema.ResourceData, m interface{}) error {
	influx := m.(meta).influxsdk
	result, err := influx.APIClient().GetScrapersID(context.Background(), &domain.GetScrapersIDAllParams{
		ScraperTargetID: d.Id(),
	})
	if err != nil {
		notFoundError := "not found: scraper target is not found"
		if err.Error() == notFoundError {
			d.SetId("")
			return nil
		}
		return fmt.Errorf("error getting Scraper: %v", err)
	}

	err = d.Set("name", *result.Name)
	if err != nil {
		return err
	}
	err = d.Set("org_id", *result.OrgID)
	if err != nil {
		return err
	}
	err = d.Set("url", *result.Url)
	if err != nil {
		return err
	}
	err = d.Set("type", *result.Type)
	if err != nil {
		return err
	}
	err = d.Set("bucket_id", *result.BucketID)
	if err != nil {
		return err
	}

	// When insecure is set to false this pointer is nil
	concreteresult := *result
	insecure := false
	if concreteresult.AllowInsecure != nil {
		insecure = *concreteresult.AllowInsecure
	}
	err = d.Set("allow_insecure", insecure)
	if err != nil {
		return err
	}

	return nil
}

func resourceScraperUpdate(d *schema.ResourceData, m interface{}) error {
	influx := m.(meta).influxsdk
	orgid := d.Get("org_id").(string)
	bucketid := d.Get("bucket_id").(string)
	name := d.Get("name").(string)
	insecure := d.Get("allow_insecure").(bool)
	targettype := domain.ScraperTargetRequestType("prometheus")
	url := d.Get("url").(string)

	updateScraper := &domain.PatchScrapersIDAllParams{
		ScraperTargetID: d.Id(),
		Body: domain.PatchScrapersIDJSONRequestBody{
			OrgID:         &orgid,
			BucketID:      &bucketid,
			Name:          &name,
			AllowInsecure: &insecure,
			Type:          &targettype,
			Url:           &url,
		},
	}
	var err error
	_, err = influx.APIClient().PatchScrapersID(context.Background(), updateScraper)

	if err != nil {
		return fmt.Errorf("error updating Scraper: %v", err)
	}

	return resourceScraperRead(d, m)
}
