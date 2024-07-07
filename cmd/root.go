package cmd

import (
	"bytes"
	"embed"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"text/template"
)

const filePerm = 0644

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:     "kratix",
	Short:   "A CLI tool for Kratix",
	Long:    `A CLI tool for Kratix`,
	Version: "v0.0.1",
	Example: `  # To initialize a new promise
  kratix init promise promise-name --group myorg.com --kind Database
`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	// Run: func(cmd *cobra.Command, args []string) { },
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	// rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.kratix-cli.yaml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func templateFiles(templates embed.FS, outputDir string, filesToTemplate map[string]string, templateValues interface{}) error {
	for path, tmpl := range filesToTemplate {
		t, err := template.ParseFS(templates, tmpl)
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
