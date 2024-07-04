package cmd

import (
	"github.com/spf13/cobra"
)

// helmPromiseCmd represents the helmPromise command
var intHelmPromiseCmd = &cobra.Command{
	Use:   "helm-promise PROMISE-NAME --url HELM-CHART-URL [--values]",
	Short: "Initialize a new Promise from a Helm chart",
	Long:  "Initialize a new Promise from a Helm Chart within the current directory, with all the necessary files to get started",
	Example: `  # initialize a new promise from the provided Helm Chart
  kratix init helm-promise postgresql --url oci://registry-1.docker.io/bitnamicharts/postgresql

  # initialize a new promise with the specified Helm chart and values.yaml
  kratix init helm-promise postgresql --url oci://registry-1.docker.io/bitnamicharts/postgresql --values values.yaml
`,
	RunE: InitHelmPromise,
	Args: cobra.ExactArgs(1),
}

var url, values string

func init() {
	initCmd.AddCommand(intHelmPromiseCmd)
	intHelmPromiseCmd.Flags().StringVarP(&url, "url", "u", "", "The URL of the Helm chart.")
	intHelmPromiseCmd.Flags().StringVarP(&values, "values", "", "", "The path to the Helm values file.")
	intHelmPromiseCmd.MarkFlagRequired("url")
}

func InitHelmPromise(cmd *cobra.Command, args []string) error {
	return templatePromiseFiles(args[0], "helm-promise")
}
