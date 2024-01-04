package influxdbv2

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/influxdata/influxdb-client-go/v2/domain"
)

func ResourceDBRPMapping() *schema.Resource {
	return &schema.Resource{
		Create: resourceDBRPMappingCreate,
		Delete: resourceDBRPMappingDelete,
		Read:   resourceDBRPMappingRead,
		Update: resourceDBRPMappingUpdate,
		Schema: map[string]*schema.Schema{
			"org_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"bucket_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"database": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"retention_policy": {
				Type:     schema.TypeString,
				Required: true,
			},
			"default_policy": {
				Type:     schema.TypeBool,
				Optional: true,
				Computed: true,
			},
		},
	}
}

func resourceDBRPMappingCreate(d *schema.ResourceData, m interface{}) error {
	influx := m.(meta).influxsdk
	bucketId := d.Get("bucket_id").(string)
	orgId := d.Get("org_id").(string)
	db := d.Get("database").(string)
	rp := d.Get("retention_policy").(string)

	dbrp, err := influx.APIClient().PostDBRP(context.Background(), &domain.PostDBRPAllParams{
		Body: domain.PostDBRPJSONRequestBody{
			BucketID:        bucketId,
			Database:        db,
			RetentionPolicy: rp,
			OrgID:           &orgId,
		}})
	if err != nil {
		return fmt.Errorf("error creating dbrp mapping: %e", err)
	}
	id := dbrp.Id

	d.SetId(id)

	return resourceDBRPMappingRead(d, m)
}

func resourceDBRPMappingDelete(d *schema.ResourceData, m interface{}) error {
	influx := m.(meta).influxsdk
	orgId := d.Get("org_id").(string)
	id := d.Id()

	err := influx.APIClient().DeleteDBRPID(context.Background(), &domain.DeleteDBRPIDAllParams{
		DbrpID: id,
		DeleteDBRPIDParams: domain.DeleteDBRPIDParams{
			OrgID: &orgId,
		},
	})

	if err != nil {
		return fmt.Errorf("error deleting dbrp: %v", err)
	}

	return nil
}

func resourceDBRPMappingRead(d *schema.ResourceData, m interface{}) error {
	influx := m.(meta).influxsdk
	orgId := d.Get("org_id").(string)

	dbrp, err := influx.APIClient().GetDBRPsID(context.Background(), &domain.GetDBRPsIDAllParams{
		DbrpID: d.Id(),
		GetDBRPsIDParams: domain.GetDBRPsIDParams{
			OrgID: &orgId,
		},
	})
	if err != nil {
		notFoundError := "not found: unable to find DBRP"
		if err.Error() == notFoundError {
			d.SetId("")
			return nil
		}
		return fmt.Errorf("error getting dbrp: %v", err)
	}

	d.SetId(*&dbrp.Content.Id)

	err = d.Set("bucket_id", dbrp.Content.BucketID)
	if err != nil {
		return err
	}
	err = d.Set("org_id", dbrp.Content.OrgID)
	if err != nil {
		return err
	}
	err = d.Set("database", dbrp.Content.Database)
	if err != nil {
		return err
	}
	err = d.Set("retention_policy", dbrp.Content.RetentionPolicy)
	if err != nil {
		return err
	}
	err = d.Set("default_policy", dbrp.Content.Default)
	if err != nil {
		return err
	}
	return nil
}

func resourceDBRPMappingUpdate(d *schema.ResourceData, m interface{}) error {
	influx := m.(meta).influxsdk
	id := d.Id()
	orgId := d.Get("org_id").(string)
	rp := d.Get("retention_policy").(string)

	_, err := influx.APIClient().PatchDBRPID(context.Background(), &domain.PatchDBRPIDAllParams{
		PatchDBRPIDParams: domain.PatchDBRPIDParams{
			OrgID: &orgId,
		},
		DbrpID: id,
		Body: domain.PatchDBRPIDJSONRequestBody{
			RetentionPolicy: &rp,
		},
	})

	if err != nil {
		return fmt.Errorf("error updating authorization: %v", err)
	}
	return resourceDBRPMappingRead(d, m)
}
