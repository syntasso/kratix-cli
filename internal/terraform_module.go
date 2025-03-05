package internal

import (
	"fmt"
	"os"
	"strings"

	"path/filepath"

	"github.com/hashicorp/go-getter"
	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclparse"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/zclconf/go-cty/cty"
)

func DownloadAndConvertTerraformToCRD(moduleSource string) ([]TerraformVariable, error) {
	tempDir, err := os.MkdirTemp("", "terraform-module")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer os.RemoveAll(tempDir)

	err = getter.Get(tempDir, moduleSource)
	if err != nil {
		return nil, fmt.Errorf("failed to download module: %w", err)
	}

	absPath := filepath.Join(tempDir, "variables.tf")

	variables, err := parseVariablesWithHCL(absPath)
	if err != nil {
		return nil, fmt.Errorf("failed to parse variables: %w", err)
	}

	return variables, nil
}

func parseVariablesWithHCL(filePath string) ([]TerraformVariable, error) {
	var variables []TerraformVariable

	// Read file content to access it later for raw source extraction
	fileBytes, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %s", err)
	}
	fileContent := string(fileBytes)

	parser := hclparse.NewParser()
	file, diags := parser.ParseHCLFile(filePath)
	if diags.HasErrors() {
		return nil, fmt.Errorf("failed to parse HCL file: %s", diags.Error())
	}

	content, diags := file.Body.Content(&hcl.BodySchema{
		Blocks: []hcl.BlockHeaderSchema{
			{
				Type:       "variable",
				LabelNames: []string{"name"},
			},
		},
	})
	if diags.HasErrors() {
		return nil, fmt.Errorf("failed to parse body content: %s", diags.Error())
	}

	for _, block := range content.Blocks {
		if block.Type != "variable" || len(block.Labels) == 0 {
			continue
		}
		name := block.Labels[0]
		variable := TerraformVariable{
			Name: name,
		}

		// Parse variable block attributes
		varContent, diags := block.Body.Content(&hcl.BodySchema{
			Attributes: []hcl.AttributeSchema{
				{Name: "type", Required: false},
				{Name: "default", Required: false},
				{Name: "description", Required: false},
			},
		})
		if diags.HasErrors() {
			// We can continue even with errors
			fmt.Printf("Warning: Some attributes for variable %s couldn't be parsed: %s\n", name, diags.Error())
		}

		// Extract type using HCL range and source extraction
		if typeAttr, ok := varContent.Attributes["type"]; ok {
			rng := typeAttr.Expr.Range()
			// Extract type from source code based on the range
			if rng.Start.Line > 0 && rng.Start.Byte < len(fileContent) && rng.End.Byte <= len(fileContent) {
				// Extract just the type expression from the file content
				typeExprText := fileContent[rng.Start.Byte:rng.End.Byte]
				variable.Type = strings.TrimSpace(typeExprText)

				// If we got a syntax expression, try to get a more specific representation
				if syntaxExpr, ok := typeAttr.Expr.(hclsyntax.Expression); ok {
					variable.Type = extractTypeFromExpr(syntaxExpr, fileContent)
				}
			}
		}

		// Extract description
		if descAttr, ok := varContent.Attributes["description"]; ok {
			descVal, diags := descAttr.Expr.Value(nil)
			if !diags.HasErrors() && descVal.Type() == cty.String {
				variable.Description = descVal.AsString()
			} else {
				// If evaluation fails, extract from source
				rng := descAttr.Expr.Range()
				if rng.Start.Line > 0 && rng.Start.Byte < len(fileContent) && rng.End.Byte <= len(fileContent) {
					descExprText := fileContent[rng.Start.Byte:rng.End.Byte]
					// For string literals, strip quotes
					if strings.HasPrefix(descExprText, "\"") && strings.HasSuffix(descExprText, "\"") {
						descExprText = descExprText[1 : len(descExprText)-1]
					}
					variable.Description = strings.TrimSpace(descExprText)
				}
			}
		}

		variables = append(variables, variable)
	}

	return variables, nil
}

// extractTypeFromExpr tries to get a more meaningful representation of a type expression
func extractTypeFromExpr(expr hclsyntax.Expression, fileContent string) string {
	switch e := expr.(type) {
	case *hclsyntax.ScopeTraversalExpr:
		// For simple types like 'string', 'number', etc.
		return e.Traversal.RootName()

	case *hclsyntax.FunctionCallExpr:
		// For function-like types like 'list(string)', 'map(number)', etc.
		args := make([]string, 0, len(e.Args))
		for _, arg := range e.Args {
			if syntaxArg, ok := arg.(hclsyntax.Expression); ok {
				args = append(args, extractTypeFromExpr(syntaxArg, fileContent))
			}
		}
		return fmt.Sprintf("%s(%s)", e.Name, strings.Join(args, ", "))

	case *hclsyntax.TemplateExpr:
		// For template expressions like "${var.something}"
		// Just return the raw source
		rng := expr.Range()
		if rng.Start.Line > 0 && rng.Start.Byte < len(fileContent) && rng.End.Byte <= len(fileContent) {
			return fileContent[rng.Start.Byte:rng.End.Byte]
		}

	default:
		// For all other types, extract from range
		rng := expr.Range()
		if rng.Start.Line > 0 && rng.Start.Byte < len(fileContent) && rng.End.Byte <= len(fileContent) {
			return fileContent[rng.Start.Byte:rng.End.Byte]
		}
	}

	return fmt.Sprintf("%T", expr)
}
