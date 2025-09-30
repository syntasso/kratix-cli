package cmd

import (
	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

// platformCmd represents the platform command
var platformCmd = &cobra.Command{
	Use:   "platform",
	Short: "A command to interact with the deployed Kratix",
	Long: `A command to interact with the deployed Kratix
	
	These sub-commands retrieve information about objects deployed in the cluster`,
}

var configFlags = genericclioptions.NewConfigFlags(true)

func init() {
	rootCmd.AddCommand(platformCmd)
	configFlags.AddFlags(platformCmd.PersistentFlags())
}
