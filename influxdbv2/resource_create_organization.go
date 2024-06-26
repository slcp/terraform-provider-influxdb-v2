package influxdbv2

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/influxdata/influxdb-client-go/v2/domain"
)

func ResourceOrganization() *schema.Resource {
	return &schema.Resource{
		Create: resourceOrganizationCreate,
		Delete: resourceOrganizationDelete,
		Read:   resourceOrganizationRead,
		Update: resourceOrganizationUpdate,
		Schema: map[string]*schema.Schema{
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"id": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"created_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"updated_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceOrganizationCreate(d *schema.ResourceData, m interface{}) error {
	influx := m.(meta).influxsdk
	desc := d.Get("description").(string)
	newOrg := &domain.Organization{
		Name:        d.Get("name").(string),
		Description: &desc,
	}
	result, err := influx.OrganizationsAPI().
		CreateOrganization(context.Background(), newOrg)
	if err != nil {
		return fmt.Errorf("error creating organization: %v", err)
	}
	d.SetId(*result.Id)
	return resourceOrganizationRead(d, m)
}

func resourceOrganizationDelete(d *schema.ResourceData, m interface{}) error {
	influx := m.(meta).influxsdk
	err := influx.OrganizationsAPI().
		DeleteOrganizationWithID(context.Background(), d.Id())
	if err != nil {
		return fmt.Errorf("error deleting organization: %v", err)
	}
	d.SetId("")
	return nil
}

func resourceOrganizationRead(d *schema.ResourceData, m interface{}) error {
	influx := m.(meta).influxsdk
	result, err := influx.OrganizationsAPI().
		FindOrganizationByID(context.Background(), d.Id())
	if err != nil {
		notFoundError := "not found: organization not found"
		if err.Error() == notFoundError {
			d.SetId("")
			return nil
		}
		return fmt.Errorf("error getting organization: %v", err)
	}
	err = d.Set("name", result.Name)
	if err != nil {
		return err
	}
	err = d.Set("description", result.Description)
	if err != nil {
		return err
	}
	err = d.Set("id", result.Id)
	if err != nil {
		return err
	}
	err = d.Set("created_at", result.CreatedAt.String())
	if err != nil {
		return err
	}
	err = d.Set("updated_at", result.UpdatedAt.String())
	if err != nil {
		return err
	}
	return nil
}

func resourceOrganizationUpdate(d *schema.ResourceData, m interface{}) error {
	influx := m.(meta).influxsdk
	id := d.Id()
	desc := d.Get("description").(string)

	updateOrg := &domain.Organization{
		Id:          &id,
		Description: &desc,
		Name:        d.Get("name").(string),
	}
	_, err := influx.OrganizationsAPI().
		UpdateOrganization(context.Background(), updateOrg)
	if err != nil {
		return fmt.Errorf("error updating organization: %v", err)
	}
	return resourceOrganizationRead(d, m)
}
