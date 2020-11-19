package secrets

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func getRequiredString(d *schema.ResourceData, name string) (string, error) {
	value, ok := d.GetOk(name)
	if !ok {
		return "", fmt.Errorf("Required  schema element '%s' not set", name)
	}

	asStr, ok := value.(string)
	if !ok {
		return "", fmt.Errorf("Schema element '%s' must be a string", name)
	}

	return asStr, nil
}
