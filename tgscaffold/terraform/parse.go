package terraform

import (
	"encoding/json"
	"fmt"

	"github.com/hashicorp/hcl2/hclwrite"
	"github.com/hashicorp/terraform-config-inspect/tfconfig"
	ctyjson "github.com/zclconf/go-cty/cty/json"
)

type Variable struct {
	Name        string
	Description string
	Type        string
	Default     *string
}

func GetModuleVariables(modulePath string) ([]Variable, error) {
	module, diags := tfconfig.LoadModule(modulePath)
	if diags.HasErrors() {
		// TODO write out diagnostics info
		return nil, diags
	}

	variables := []Variable{}
	for _, variable := range module.Variables {
		defaultStringRep, err := moduleVariableToHCLString(variable.Default)
		if err != nil {
			return nil, err
		}

		variables = append(
			variables,
			Variable{
				Name:        variable.Name,
				Description: variable.Description,
				Type:        variable.Type,
				Default:     defaultStringRep,
			},
		)
	}
	return variables, nil
}

func moduleVariableToHCLString(defaultVal interface{}) (*string, error) {
	if defaultVal == nil {
		return nil, nil
	}

	// To get the HCL representation of the default value, we need to first convert the interface to json. Otherwise, we
	// can't get the types!
	defaultValAsJson, err := json.Marshal(defaultVal)
	if err != nil {
		return nil, err
	}

	// We then convert the json val to cty
	impliedType, err := ctyjson.ImpliedType(defaultValAsJson)
	if err != nil {
		return nil, err
	}
	defaultValAsCty, err := ctyjson.Unmarshal(defaultValAsJson, impliedType)
	if err != nil {
		return nil, err
	}

	// Get the HCL2 representation of the cty value
	tokens := hclwrite.TokensForValue(defaultValAsCty)
	fout := hclwrite.NewEmptyFile()
	rootBody := fout.Body()
	rootBody.AppendUnstructuredTokens(tokens)

	out := fmt.Sprintf("%s", fout.Bytes())
	return &out, nil
}
