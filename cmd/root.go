package cmd

import (
	"bytes"
	"embed"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/Masterminds/sprig/v3"
	"github.com/spf13/cobra"
)

const filePerm = 0644

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:     "kratix",
	Short:   "A CLI tool for Kratix",
	Long:    `A CLI tool for Kratix`,
	Version: "",
	Example: `  # To initialize a new promise
  kratix init promise promise-name --group myorg.com --kind Database
`,
}

func Execute(version string) {
	rootCmd.Version = version
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func templateFiles(templates embed.FS, outputDir string, filesToTemplate map[string]string, templateValues interface{}) error {
	for path, tmpl := range filesToTemplate {
		t, err := template.New(filepath.Base(tmpl)).Funcs(sprig.FuncMap()).ParseFS(templates, tmpl)
		if err != nil {
			return err
		}

		data := bytes.NewBuffer([]byte{})
		if err := t.Execute(data, templateValues); err != nil {
			return err
		}
		fullPath := filepath.Join(outputDir, path)
		parentDir := filepath.Dir(fullPath)
		if _, err := os.Stat(parentDir); os.IsNotExist(err) {
			if err := os.MkdirAll(parentDir, os.ModePerm); err != nil {
				return err
			}
		}

		if err := os.WriteFile(fullPath, data.Bytes(), filePerm); err != nil {
			return err
		}
	}
	return nil
}

func ParseContainerCmdArgs(containerPath string) (*ContainerCmdArgs, error) {
	parts := strings.Split(containerPath, "/")
	if len(parts) != 3 {
		return nil, fmt.Errorf("invalid pipeline format: %s, expected format: LIFECYCLE/ACTION/PIPELINE-NAME", containerPath)
	}

	return &ContainerCmdArgs{
		Lifecycle: parts[0],
		Action:    parts[1],
		Pipeline:  parts[2],
	}, nil
}
