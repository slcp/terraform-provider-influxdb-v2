package influxdbv2

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func ResourceLegacyAuthorization() *schema.Resource {
	return &schema.Resource{
		Create: resourceLegacyAuthorizationCreate,
		Delete: resourceLegacyAuthorizationDelete,
		Read:   resourceLegacyAuthorizationRead,
		Update: resourceLegacyAuthorizationUpdate,
		Schema: map[string]*schema.Schema{
			"org_id": {
				Type:     schema.TypeString,
				ForceNew: true,
				Required: true,
			},
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"status": {
				Type:     schema.TypeString,
				Optional: true,
				Default:  "active",
			},
			"name": {
				Type:     schema.TypeString,
				ForceNew: true,
				Required: true,
			},
			"password": {
				Type:      schema.TypeString,
				Required:  true,
				Sensitive: true,
			},
			"permissions": {
				Type:     schema.TypeSet,
				Required: true,
				ForceNew: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"action": {
							Type:     schema.TypeString,
							ForceNew: true,
							Required: true,
						},
						"resource": {
							Type:     schema.TypeSet,
							ForceNew: true,
							Required: true,
							Elem: &schema.Resource{
								Schema: map[string]*schema.Schema{
									"id": {
										Type:     schema.TypeString,
										ForceNew: true,
										Required: true,
									},
									"org": {
										Type:     schema.TypeString,
										ForceNew: true,
										Optional: true,
									},
									"org_id": {
										Type:     schema.TypeString,
										ForceNew: true,
										Required: true,
									},
									"type": {
										Type:     schema.TypeString,
										ForceNew: true,
										Required: true,
									},
								},
							},
						},
					},
				},
			},
			"user_id": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Computed: true,
			},
			"user_org_id": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
				Computed: true,
			},
		},
	}
}

func resourceLegacyAuthorizationCreate(d *schema.ResourceData, m interface{}) error {
	influx := m.(meta).legacyAuthorizationsClient
	description := d.Get("description").(string)
	orgId := d.Get("org_id").(string)
	username := d.Get("name").(string)
	password := d.Get("password").(string)
	status := LegacyAuthorizationPostRequestStatus(d.Get("status").(string))
	permissions := getLegacyPermissions(d.Get("permissions"))
	ctx := context.Background()

	// Create an authorization
	authorization, err := influx.PostLegacyAuthorizationsWithResponse(ctx, &PostLegacyAuthorizationsParams{}, PostLegacyAuthorizationsJSONRequestBody{
		Description: &description,
		Token:       &username,
		Permissions: &permissions,
		OrgID:       &orgId,
		Status:      &status,
	})
	if err != nil || authorization.StatusCode() != 201 {
		resourceLegacyAuthorizationDelete(d, m)
		return fmt.Errorf("error creating legacy authorization: %e", err)
	}
	userId := *authorization.JSON201.Id

	// Add a password to the authorization
	pass, err := influx.PostLegacyAuthorizationsIDPasswordWithResponse(ctx, userId, &PostLegacyAuthorizationsIDPasswordParams{}, PostLegacyAuthorizationsIDPasswordJSONRequestBody{
		Password: password,
	})
	// If password fails, delete the authorization
	if err != nil || pass.StatusCode() != 204 {
		resourceLegacyAuthorizationDelete(d, m)
		return fmt.Errorf("error creating legacy authorization password: %e", err)
	}

	d.SetId(userId)
	err = d.Set("name", *authorization.JSON201.Token)
	if err != nil {
		return err
	}
	if err != nil {
		return err
	}
	return resourceLegacyAuthorizationRead(d, m)
}

func resourceLegacyAuthorizationDelete(d *schema.ResourceData, m interface{}) error {
	influx := m.(meta).legacyAuthorizationsClient
	id := d.Id()
	_, err := influx.DeleteLegacyAuthorizationsIDWithResponse(context.Background(), id, &DeleteLegacyAuthorizationsIDParams{})
	if err != nil {
		return fmt.Errorf("error deleting authorization: %v", err)
	}
	return nil
}

func resourceLegacyAuthorizationRead(d *schema.ResourceData, m interface{}) error {
	influx := m.(meta).legacyAuthorizationsClient
	password := d.Get("password").(string)

	authorization, err := influx.GetLegacyAuthorizationsIDWithResponse(context.Background(), d.Id(), &GetLegacyAuthorizationsIDParams{})
	if err != nil || authorization.StatusCode() != 200 {
		return fmt.Errorf("error getting authorization: %v", err)
	}

	err = d.Set("status", authorization.JSON200.Status)
	if err != nil {
		return err
	}
	err = d.Set("user_id", authorization.JSON200.UserID)
	if err != nil {
		return err
	}
	err = d.Set("user_org_id", authorization.JSON200.OrgID)
	if err != nil {
		return err
	}
	err = d.Set("name", authorization.JSON200.Token)
	if err != nil {
		return err
	}
	err = d.Set("org_id", authorization.JSON200.OrgID)
	if err != nil {
		return err
	}
	err = d.Set("description", authorization.JSON200.Description)
	if err != nil {
		return err
	}
	err = d.Set("password", password)
	if err != nil {
		return err
	}
	return nil
}

func resourceLegacyAuthorizationUpdate(d *schema.ResourceData, m interface{}) error {
	influx := m.(meta).legacyAuthorizationsClient
	id := d.Id()
	description := d.Get("description").(string)
	password := d.Get("password").(string)
	status := AuthorizationUpdateRequestStatus(d.Get("status").(string))
	ctx := context.Background()

	authorization, err := influx.PatchLegacyAuthorizationsIDWithResponse(ctx, id, &PatchLegacyAuthorizationsIDParams{}, PatchLegacyAuthorizationsIDJSONRequestBody{
		Description: &description,
		Status:      &status,
	})
	if err != nil || authorization.StatusCode() != 200 {
		return fmt.Errorf("error updating legacy authorization: %e", err)
	}

	// Update the password on the authorization
	pass, err := influx.PostLegacyAuthorizationsIDPasswordWithResponse(ctx, id, &PostLegacyAuthorizationsIDPasswordParams{}, PostLegacyAuthorizationsIDPasswordJSONRequestBody{
		Password: password,
	})
	if err != nil || pass.StatusCode() != 204 {
		return fmt.Errorf("error updating legacy authorization password: %e", err)
	}

	return resourceLegacyAuthorizationRead(d, m)
}

func getLegacyPermissions(input interface{}) []Permission {
	result := []Permission{}
	permissionsSet := input.(*schema.Set).List()
	for _, permission := range permissionsSet {
		perm, ok := permission.(map[string]interface{})
		if ok {
			resourceSet := perm["resource"].(*schema.Set).List()
			for _, resource := range resourceSet {
				res := resource.(map[string]interface{})
				var id, org_id, name, org = "", "", "", ""
				if res["id"] != nil {
					id = res["id"].(string)
				}
				if res["org_id"] != nil {
					org_id = res["org_id"].(string)
				}
				if res["name"] != nil {
					name = res["name"].(string)
				}
				if res["org"] != nil {
					org = res["org"].(string)
				}
				Resource := Resource{Type: ResourceType(res["type"].(string)), Id: &id, OrgID: &org_id, Name: &name, Org: &org}
				each := Permission{Action: PermissionAction(perm["action"].(string)), Resource: Resource}
				result = append(result, each)
			}
		}
	}
	return result
}

func getLegacyAuthorizationsById(input *[]Authorization, id string) Authorization {
	result := Authorization{}
	for _, authorization := range *input {
		if *authorization.Id == id {
			result = authorization
		}
	}
	return result
}
