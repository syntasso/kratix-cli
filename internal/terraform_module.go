package internal

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclparse"
	"github.com/hashicorp/hcl/v2/hclsyntax"
	"github.com/zclconf/go-cty/cty"
)

const kratixModuleName = "kratix_target"

type terraformModuleManifest struct {
	Modules []struct {
		Key string `json:"Key"`
		Dir string `json:"Dir"`
	} `json:"Modules"`
}

var (
	mkdirTemp     func(dir, pattern string) (string, error) = os.MkdirTemp
	terraformInit func(dir string) error                    = runTerraformInit
)

func GetVariablesFromModule(moduleSource, moduleRegistryVersion string) ([]TerraformVariable, error) {
	tempDir, err := mkdirTemp("", "terraform-module")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer os.RemoveAll(tempDir)

	if err := writeTerraformModuleConfig(tempDir, moduleSource, moduleRegistryVersion); err != nil {
		return nil, err
	}

	if err := terraformInit(tempDir); err != nil {
		return nil, fmt.Errorf("failed to initialize terraform: %w", err)
	}

	moduleDir, err := resolveModuleDir(tempDir)
	if err != nil {
		return nil, err
	}

	absPath := filepath.Join(moduleDir, "variables.tf")
	variables, err := extractVariablesFromVarsFile(absPath)
	if err != nil {
		return nil, fmt.Errorf("failed to parse variables: %w", err)
	}

	return variables, nil
}

func writeTerraformModuleConfig(workDir, moduleSource, moduleRegistryVersion string) error {
	config := fmt.Sprintf("module \"%s\" {\n  source = \"%s\"\n", kratixModuleName, moduleSource)
	if moduleRegistryVersion != "" {
		config += fmt.Sprintf("  version = \"%s\"\n", moduleRegistryVersion)
	}
	config += "}\n"
	if err := os.WriteFile(filepath.Join(workDir, "main.tf"), []byte(config), 0o644); err != nil {
		return fmt.Errorf("failed to write terraform config: %w", err)
	}

	return nil
}

func runTerraformInit(dir string) error {
	cmd := exec.Command("terraform", "init", "-backend=false")
	cmd.Dir = dir

	if output, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("terraform init failed: %w: %s", err, string(output))
	}

	return nil
}

func resolveModuleDir(workDir string) (string, error) {
	manifestPath := filepath.Join(workDir, ".terraform", "modules", "modules.json")
	manifestBytes, err := os.ReadFile(manifestPath)
	if err != nil {
		return "", fmt.Errorf("failed to read terraform module manifest: %w", err)
	}

	var manifest terraformModuleManifest
	if err := json.Unmarshal(manifestBytes, &manifest); err != nil {
		return "", fmt.Errorf("failed to unmarshal terraform module manifest: %w", err)
	}

	for _, module := range manifest.Modules {
		moduleKey := strings.TrimPrefix(module.Key, "module.")
		if moduleKey != kratixModuleName {
			continue
		}

		if filepath.IsAbs(module.Dir) {
			return filepath.Clean(module.Dir), nil
		}

		return filepath.Clean(filepath.Join(workDir, module.Dir)), nil
	}

	return "", fmt.Errorf("module %s not found in terraform module manifest", kratixModuleName)
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
		d, err := extractDefault(varContent)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error extracting default for variable %s: %s\n", variable.Name, err)
			fmt.Fprintln(os.Stderr, "Continuing without default value")
		}
		variable.Default = d
		variables = append(variables, variable)
	}

	return variables
}

func extractDefault(varContent *hcl.BodyContent) (any, error) {
	if defaultAttr, ok := varContent.Attributes["default"]; ok {
		defaultVal, diags := defaultAttr.Expr.Value(nil)
		if diags.HasErrors() {
			return nil, fmt.Errorf("failed to evaluate default value: %s", diags.Error())
		}

		if defaultVal.IsNull() {
			return nil, nil
		}
		return convertCtyValue(defaultVal), nil
	}

	return nil, nil
}

func convertCtySliceToGoSlice(values []cty.Value) any {
	if len(values) == 0 {
		return []any{}
	}

	firstType := values[0].Type()
	switch firstType {
	case cty.String:
		result := make([]string, len(values))
		for i, v := range values {
			result[i] = v.AsString()
		}
		return result

	case cty.Number:
		result := make([]float64, len(values))
		for i, v := range values {
			f, _ := v.AsBigFloat().Float64()
			result[i] = f
		}
		return result

	case cty.Bool:
		result := make([]bool, len(values))
		for i, v := range values {
			result[i] = v.True()
		}
		return result

	default:
		result := make([]any, len(values))
		for i, v := range values {
			result[i] = convertCtyValue(v)
		}
		return result
	}
}

func convertCtyMapToGoMap(values map[string]cty.Value) map[string]any {
	result := make(map[string]any)
	for k, v := range values {
		result[k] = convertCtyValue(v)
	}
	return result
}

func convertCtyObjectToGoObject(values map[string]cty.Value) map[string]any {
	result := make(map[string]any)
	for k, v := range values {
		result[k] = convertCtyValue(v)
	}
	return result
}

func convertCtyValue(v cty.Value) any {
	switch {
	case v.Type() == cty.String:
		return v.AsString()
	case v.Type() == cty.Number:
		f, _ := v.AsBigFloat().Float64()
		return f
	case v.Type() == cty.Bool:
		return v.True()
	case v.Type().IsListType() || v.Type().IsTupleType():
		return convertCtySliceToGoSlice(v.AsValueSlice())
	case v.Type().IsMapType():
		return convertCtyMapToGoMap(v.AsValueMap())
	case v.Type().IsObjectType():
		return convertCtyObjectToGoObject(v.AsValueMap())
	default:
		return v
	}
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
			args = append(args, extractTypeFromExpr(arg, fileContent))
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
