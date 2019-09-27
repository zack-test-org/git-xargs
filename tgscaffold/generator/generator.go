// Package containing functions and routines for rendering the templates to generate the terragrunt config
package generator

import (
	"os"
	"path/filepath"
	"text/template"

	"github.com/Masterminds/sprig"
	"github.com/gobuffalo/packr/v2"
	"github.com/gruntwork-io/prototypes/tgscaffold/terraform"
)

// This path is relative to the source file, and only important during local development. The relativity is ignored when
// built into packr2 and will automatically source from the binary.
const TemplateRootPath = "../templates"

var TemplatePaths = []string{
	"terragrunt.hcl",
}

type RenderOptions struct {
	Constants ConstantsType

	Cloud string

	InfrastructureModulesSource  string
	PathToModule                 string
	InfrastructureModulesVersion string

	RequiredVars []terraform.Variable
	OptionalVars []terraform.Variable
}

func Render(options RenderOptions, targetDir string) error {
	if err := os.MkdirAll(targetDir, os.ModePerm); err != nil {
		return err
	}

	box := packr.New("tgscaffold", TemplateRootPath)
	for _, templatePath := range TemplatePaths {
		tpl, err := box.FindString(templatePath + ".tpl")
		if err != nil {
			return err
		}

		parsed, err := template.New(templatePath).Funcs(sprig.TxtFuncMap()).Parse(tpl)
		if err != nil {
			return err
		}
		if err = renderTemplateToFile(parsed, filepath.Join(targetDir, templatePath), options); err != nil {
			return err
		}
	}
	return nil
}

func renderTemplateToFile(tpl *template.Template, outFname string, data interface{}) error {
	outF, err := os.Create(outFname)
	defer outF.Close()
	if err != nil {
		return err
	}
	return tpl.Execute(outF, data)
}
