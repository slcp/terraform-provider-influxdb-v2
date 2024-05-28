package influxdbv2

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/influxdata/influxdb-client-go/v2/domain"
)

func ResourceBucket() *schema.Resource {
	return &schema.Resource{
		Create: resourceBucketCreate,
		Delete: resourceBucketDelete,
		Read:   resourceBucketRead,
		Update: resourceBucketUpdate,
		Schema: map[string]*schema.Schema{
			"description": {
				Type:     schema.TypeString,
				Optional: true,
			},
			"name": {
				Type:     schema.TypeString,
				Required: true,
			},
			"org_id": {
				Type:     schema.TypeString,
				Required: true,
				ForceNew: true,
			},
			"retention_rules": {
				Type:     schema.TypeSet,
				Required: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"every_seconds": {
							Type:     schema.TypeInt,
							Required: true,
						},
						"shard_group_duration_seconds": {
							Type:     schema.TypeInt,
							Optional: true,
						},
						"type": {
							Type:     schema.TypeString,
							Optional: true,
							Default:  "expire",
						},
					},
				},
			},
			"rp": {
				Type:     schema.TypeString,
				Optional: true,
				ForceNew: true,
			},
			"created_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"updated_at": {
				Type:     schema.TypeString,
				Computed: true,
			},
			"type": {
				Type:     schema.TypeString,
				Computed: true,
			},
		},
	}
}

func resourceBucketCreate(d *schema.ResourceData, m interface{}) error {
	influx := m.(meta).influxsdk

	retentionRules, err := getRetentionRules(d.Get("retention_rules"))
	if err != nil {
		return err
	}

	desc := d.Get("description").(string)
	orgid := d.Get("org_id").(string)
	retpe := d.Get("rp").(string)
	newBucket := &domain.Bucket{
		Description:    &desc,
		Name:           d.Get("name").(string),
		OrgID:          &orgid,
		RetentionRules: retentionRules,
		Rp:             &retpe,
	}
	result, err := influx.BucketsAPI().CreateBucket(context.Background(), newBucket)
	if err != nil {
		return fmt.Errorf("error creating bucket: %v", err)
	}
	d.SetId(*result.Id)
	return resourceBucketRead(d, m)
}

func resourceBucketDelete(d *schema.ResourceData, m interface{}) error {
	influx := m.(meta).influxsdk
	err := influx.BucketsAPI().DeleteBucketWithID(context.Background(), d.Id())
	if err != nil {
		return fmt.Errorf("error deleting bucket: %v", err)
	}
	d.SetId("")
	return nil
}

func resourceBucketRead(d *schema.ResourceData, m interface{}) error {
	influx := m.(meta).influxsdk

	// Get user provided retention rules
	providedRR := d.Get("retention_rules").(*schema.Set).List()

	result, err := influx.BucketsAPI().FindBucketByID(context.Background(), d.Id())
	if err != nil {
		notFoundError := "not found: bucket not found"
		if err.Error() == notFoundError {
			d.SetId("")
			return nil
		}
		return fmt.Errorf("error getting bucket: %v", err)
	}

	// Reformat retention rules array
	var rr []map[string]interface{}
	for i, retention_rule := range result.RetentionRules {
		tmp := map[string]interface{}{
			"every_seconds": retention_rule.EverySeconds,
			"type":          retention_rule.Type,
		}
		// Get user provided retention rules
		provided, ok := providedRR[i].(map[string]interface{})
		if !ok {
			return fmt.Errorf("error reading user provided retention rules")
		}
		// If the user provided shard_group_duration_seconds, return the value from the API so that
		// Terraform can produce a plan against it. If not, don't return because this is an optional
		// property. It shouldn't be present if not declared. When creating/updating we make sure it
		// is set to the default value when not provided by the user.
		if provided["shard_group_duration_seconds"].(int) > 0 {
			tmp["shard_group_duration_seconds"] = int(*retention_rule.ShardGroupDurationSeconds)
		}
		if provided["shard_group_duration_seconds"].(int) == -1 {
			tmp["shard_group_duration_seconds"] = -1
		}
		rr = append(rr, tmp)
	}
	if len(result.RetentionRules) == 0 {
		// If no retention rules, there's a default of 0 expiry
		// but this isn't returned on the API
		tmp := map[string]interface{}{
			"every_seconds": 0,
			"type":          "expire",
		}
		rr = append(rr, tmp)
	}

	err = d.Set("name", result.Name)
	if err != nil {
		return err
	}
	err = d.Set("description", result.Description)
	if err != nil {
		return err
	}
	err = d.Set("org_id", result.OrgID)
	if err != nil {
		return err
	}
	err = d.Set("retention_rules", rr)
	if err != nil {
		return err
	}
	err = d.Set("rp", result.Rp)
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
	err = d.Set("type", result.Type)
	if err != nil {
		return err
	}

	return nil
}

func resourceBucketUpdate(d *schema.ResourceData, m interface{}) error {
	var err error
	influx := m.(meta).influxsdk

	retentionRules, err := getRetentionRules(d.Get("retention_rules"))
	if err != nil {
		return err
	}

	id := d.Id()
	desc := d.Get("description").(string)
	orgid := d.Get("org_id").(string)
	retpe := d.Get("rp").(string)
	updateBucket := &domain.Bucket{
		Id:             &id,
		Description:    &desc,
		Name:           d.Get("name").(string),
		OrgID:          &orgid,
		RetentionRules: retentionRules,
		Rp:             &retpe,
	}
	_, err = influx.BucketsAPI().UpdateBucket(context.Background(), updateBucket)

	if err != nil {
		return fmt.Errorf("error updating bucket: %v", err)
	}

	return resourceBucketRead(d, m)
}

func getRetentionRules(input interface{}) (domain.RetentionRules, error) {
	result := domain.RetentionRules{}
	retentionRulesSet := input.(*schema.Set).List()
	for _, retentionRule := range retentionRulesSet {
		defaultType := domain.RetentionRuleType("expire")

		rr, ok := retentionRule.(map[string]interface{})
		if !ok {
			return nil, fmt.Errorf("Error reading retention rules")
		}
		each := domain.RetentionRule{
			EverySeconds: int64(rr["every_seconds"].(int)),
			Type:         &defaultType,
		}

		seconds, ok := rr["shard_group_duration_seconds"].(int)
		if !ok {
			return nil, fmt.Errorf("Error reading retention rules duration")
		}

		// -1 is a flag that signals the user does not wish for the shard group
		// duration to be managed in any way by the provider
		if seconds == -1 {
			result = append(result, each)
			continue
		}

		if seconds > 0 {
			v := int64(seconds)
			each.ShardGroupDurationSeconds = &v
			result = append(result, each)
			continue
		}

		// When user has not provided a shard group duration, ensure that
		// it set to the default as we could be updating instead of creating
		d := getDefaultShardGroupDuration(each.EverySeconds)
		each.ShardGroupDurationSeconds = &d
		result = append(result, each)
	}
	return result, nil
}

// https://docs.influxdata.com/influxdb/v2/reference/internals/shards/#shard-group-duration
func getDefaultShardGroupDuration(rps int64) int64 {
	hour := int64(60 * 60)
	day := hour * 24
	month := 180 * day
	if rps < 2*day {
		return 1 * hour
	}
	if rps < 6*month {
		return 1 * day
	}
	return 7 * day
}
