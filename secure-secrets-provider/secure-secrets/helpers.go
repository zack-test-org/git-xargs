package secure_secrets

import (
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"reflect"
)

func getRequiredString(d *schema.ResourceData, name string) (string, error) {
	value, ok := d.GetOk(name)
	if !ok {
		return "", fmt.Errorf("Required  schema element '%s' not set", name)
	}

	asStr, ok := value.(string)
	if !ok {
		return "", fmt.Errorf("Schema element '%s' must be a string, but got '%v'", name, reflect.TypeOf(value))
	}

	return asStr, nil
}

func getOptionalString(d *schema.ResourceData, name string) (*string, error) {
	value, ok := d.GetOk(name)
	if !ok {
		return nil, nil
	}

	asStr, ok := value.(string)
	if !ok {
		return nil, fmt.Errorf("Schema element '%s' must be a string, but got '%v'", name, reflect.TypeOf(value))
	}

	return &asStr, nil
}
