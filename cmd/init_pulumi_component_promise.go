package cmd

import (
	"github.com/spf13/cobra"
	"github.com/syntasso/kratix-cli/internal/pulumi"
)

const (
	pulumiComponentPromiseCommandName = "pulumi-component-promise"
	pulumiComponentPromiseExamples    = `  # initialize a new promise from a local Pulumi package schema
  kratix init pulumi-component-promise mypromise --schema ./schema.json --group syntasso.io --kind Database

  # initialize a new promise from a remote Pulumi package schema
  kratix init pulumi-component-promise mypromise --schema https://www.pulumi.com/registry/packages/aws/schema.json --group syntasso.io --kind Database
`
)

var (
	pulumiSchemaPath string
	pulumiComponent  string
)

var pulumiComponentPromiseCmd = &cobra.Command{
	Use:   pulumiComponentPromiseCommandName + " PROMISE-NAME --schema PATH_OR_URL --group PROMISE-API-GROUP --kind PROMISE-API-KIND [--component TOKEN] [--version] [--plural] [--split] [--dir DIR]",
	Short: "Preview: Initialize a new Promise from a Pulumi package schema",
	Long: "Preview: Initialize a new Promise from a Pulumi package schema. " +
		"This command is in preview, not supported under SLAs, and may change or break without notice.",
	Example: pulumiComponentPromiseExamples,
	Args:    cobra.ExactArgs(1),
	RunE:    InitPulumiComponentPromise,
}

func init() {
	initCmd.AddCommand(pulumiComponentPromiseCmd)

	pulumiComponentPromiseCmd.Flags().StringVar(&pulumiSchemaPath, "schema", "", "Path or URL to Pulumi package schema")
	pulumiComponentPromiseCmd.Flags().StringVar(&pulumiComponent, "component", "", "Pulumi component token to use from the schema")

	pulumiComponentPromiseCmd.MarkFlagRequired("schema")
}

func InitPulumiComponentPromise(cmd *cobra.Command, args []string) error {
	printPreviewWarning()
	schemaDoc, err := pulumi.LoadSchema(pulumiSchemaPath)
	if err != nil {
		return err
	}

	selectedComponent, err := pulumi.SelectComponent(schemaDoc, pulumiComponent)
	if err != nil {
		return err
	}

	return initPulumiComponentPromiseFromSelection(selectedComponent)
}

func initPulumiComponentPromiseFromSelection(component pulumi.SelectedComponent) error {
	_ = component // selected component is intentionally passed forward for translation in the next task.
	return nil
}
