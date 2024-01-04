package influxdbv2

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
)

func extractIdForResource(s *terraform.State, name string) string {
	is := findResourceInState(s, name)
	v, _ := is.Attributes["id"]
	return v
}

func findResourceInState(s *terraform.State, name string) *terraform.InstanceState {
	ms := s.RootModule()
	rs, ok := ms.Resources[name]
	if !ok {
		return nil
	}
	is := rs.Primary
	return is
}

func checkResourceHasBeenReplaced(name string, oid *string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		if id := extractIdForResource(s, name); id == *oid {
			return fmt.Errorf("id should have changed but it is still %s", id)
		}
		return nil
	}
}
