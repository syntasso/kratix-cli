package cmd

import (
	"github.com/spf13/cobra"
)

var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Command used to initialize Kratix resources",
	Long:  `Command used to initialize Kratix resources"`,
}

var (
	group, kind, version, plural, outputDir string
)

func init() {
	rootCmd.AddCommand(initCmd)

	initCmd.PersistentFlags().StringVarP(&group, "group", "g", "", "The API group for the Promise")
	initCmd.PersistentFlags().StringVarP(&kind, "kind", "k", "", "The kind to be provided by the Promise")
	initCmd.PersistentFlags().StringVarP(&version, "version", "v", "", "The group version for the Promise. Defaults to v1alpha1")
	initCmd.PersistentFlags().StringVar(&plural, "plural", "", "The plural form of the kind. Defaults to the kind name with an additional 's' at the end.")
	initCmd.PersistentFlags().StringVarP(&outputDir, "dir", "d", ".", "The output directory to write the Promise structure to; defaults to '.'")

	initCmd.MarkPersistentFlagRequired("group")
	initCmd.MarkPersistentFlagRequired("kind")
}
