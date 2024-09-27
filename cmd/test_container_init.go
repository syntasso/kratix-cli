package cmd

import (
	"embed"
	"fmt"
	"os"
	"path"
	"strings"

	"github.com/spf13/cobra"
)

//go:embed templates/test/*
var testTemplates embed.FS

var testContainerInitCmd = &cobra.Command{
	Use:   "init --image CONTAINER-IMAGE",
	Short: "Command to initialise a testcase directory for a container image",
	Example: `  # initialise a container test directory for a given image
  > kratix test container init --image ghcr.io/syntasso/my-image:v0.1.0`,
	RunE: TestContainerInit,
	Args: cobra.ExactArgs(0),
}

type testTemplateValues struct {
	RawImageName       string
	FormattedImageName string
	Directory          string
}

func init() {
	testContainerCmd.AddCommand(testContainerInitCmd)
}

func TestContainerInit(cmd *cobra.Command, args []string) error {
	imageName := strings.Split(testImage, ":")[0]
	formattedImageName := strings.ReplaceAll(imageName, "/", "_")
	formattedImageName = strings.ReplaceAll(formattedImageName, ".", "_")

	imageDir := path.Join(testcaseDir, formattedImageName)

	if err := os.MkdirAll(imageDir, 0755); err != nil {
		return err
	}

	templates := map[string]string{
		"README.md": "templates/test/README.md.tpl",
	}

	templateValues := testTemplateValues{
		RawImageName:       imageName,
		FormattedImageName: formattedImageName,
		Directory:          imageDir,
	}

	if err := templateFiles(testTemplates, testcaseDir, templates, templateValues); err != nil {
		return err
	}

	fmt.Println("Initialised container test directory:")
	fmt.Println("  " + imageDir)

	return nil
}
