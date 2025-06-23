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

var (
	getModule func(dst, src string, opts ...getter.ClientOption) error = getter.Get
	mkdirTemp func(dir, pattern string) (string, error)                = os.MkdirTemp
)

func GetVariablesFromModule(moduleSource string) ([]TerraformVariable, error) {
	tempDir, err := mkdirTemp("", "terraform-module")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer os.RemoveAll(tempDir)

	err = getModule(tempDir, moduleSource)
	if err != nil {
		return nil, fmt.Errorf("failed to download module: %w", err)
	}

	absPath := filepath.Join(tempDir, "variables.tf")

	variables, err := extractVariablesFromVarsFile(absPath)
	if err != nil {
		return nil, fmt.Errorf("failed to parse variables: %w", err)
	}

	return variables, nil
}

func extractVariablesFromVarsFile(filePath string) ([]TerraformVariable, error) {
	fileContent, err := readFileContent(filePath)
	if err != nil {
		return nil, err
	}

	blocks, err := parseHCLVariables(filePath)
	if err != nil {
		return nil, err
	}

	return extractVariables(blocks, fileContent), nil
}

func readFileContent(filePath string) (string, error) {
	fileBytes, err := os.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to read file: %s", err)
	}
	return string(fileBytes), nil
}

func parseHCLVariables(filePath string) ([]*hcl.Block, error) {
	parser := hclparse.NewParser()
	file, diags := parser.ParseHCLFile(filePath)
	if diags.HasErrors() {
		return nil, fmt.Errorf("failed to parse HCL file: %s", diags.Error())
	}

	content, diags := file.Body.Content(&hcl.BodySchema{
		Blocks: []hcl.BlockHeaderSchema{
			{Type: "variable", LabelNames: []string{"name"}},
		},
	})
	if diags.HasErrors() {
		return nil, fmt.Errorf("failed to parse body content: %s", diags.Error())
	}

	return content.Blocks, nil
}

func extractVariables(blocks []*hcl.Block, fileContent string) []TerraformVariable {
	var variables []TerraformVariable

	for _, block := range blocks {
		if block.Type != "variable" || len(block.Labels) == 0 {
			continue
		}
		variable := TerraformVariable{Name: block.Labels[0]}
		varContent, _ := block.Body.Content(&hcl.BodySchema{
			Attributes: []hcl.AttributeSchema{
				{Name: "type", Required: false},
				{Name: "default", Required: false},
				{Name: "description", Required: false},
			},
		})

		variable.Type = extractType(varContent, fileContent)
		variable.Description = extractDescription(varContent, fileContent)
		d, err := extractDefault(varContent, fileContent)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error extracting default for variable %s: %s\n", variable.Name, err)
			fmt.Fprintln(os.Stderr, "Continuing without default value")
		}
		variable.Default = d
		variables = append(variables, variable)
	}

	return variables
}

func extractDefault(varContent *hcl.BodyContent, fileContent string) (any, error) {
	if defaultAttr, ok := varContent.Attributes["default"]; ok {
		defaultVal, diags := defaultAttr.Expr.Value(nil)
		if diags.HasErrors() {
			return nil, fmt.Errorf("failed to evaluate default value: %s", diags.Error())
		}

		if defaultVal.IsNull() {
			return nil, nil
		}
		if defaultVal.Type() == cty.String {
			return defaultVal.AsString(), nil
		}
		if defaultVal.Type() == cty.Number {
			return defaultVal.AsBigFloat().String(), nil
		}
		if defaultVal.Type() == cty.Bool {
			return defaultVal.True(), nil
		}
		if defaultVal.Type().IsTupleType() || defaultVal.Type().IsListType() || defaultVal.Type().IsMapType() {
			return defaultVal.AsValueSlice(), nil
		}

		return defaultVal, nil
	}

	return nil, nil
}

func extractType(varContent *hcl.BodyContent, fileContent string) string {
	if typeAttr, ok := varContent.Attributes["type"]; ok {
		rng := typeAttr.Expr.Range()
		if rng.Start.Line > 0 && rng.Start.Byte < len(fileContent) && rng.End.Byte <= len(fileContent) {
			typeExprText := fileContent[rng.Start.Byte:rng.End.Byte]
			if syntaxExpr, ok := typeAttr.Expr.(hclsyntax.Expression); ok {
				return extractTypeFromExpr(syntaxExpr, fileContent)
			}
			return strings.TrimSpace(typeExprText)
		}
	}
	return ""
}

func extractDescription(varContent *hcl.BodyContent, fileContent string) string {
	if descAttr, ok := varContent.Attributes["description"]; ok {
		descVal, diags := descAttr.Expr.Value(nil)
		if !diags.HasErrors() && descVal.Type() == cty.String {
			return descVal.AsString()
		}
		rng := descAttr.Expr.Range()
		if rng.Start.Line > 0 && rng.Start.Byte < len(fileContent) && rng.End.Byte <= len(fileContent) {
			descExprText := fileContent[rng.Start.Byte:rng.End.Byte]
			if strings.HasPrefix(descExprText, "\"") && strings.HasSuffix(descExprText, "\"") {
				return strings.TrimSpace(descExprText[1 : len(descExprText)-1])
			}
			return strings.TrimSpace(descExprText)
		}
	}
	return ""
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
