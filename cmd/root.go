package cmd

import (
	"bytes"
	"embed"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/Masterminds/sprig/v3"
	"github.com/go-logr/stdr"
	"github.com/spf13/cobra"
	ctrl "sigs.k8s.io/controller-runtime"
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
	ctrl.SetLogger(stdr.New(log.New(os.Stderr, "", log.LstdFlags)))
	if err := handlePotentialPluginCommand(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func templateFiles(templates embed.FS, outputDir string, filesToTemplate map[string]string, templateValues any) error {
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

func handlePotentialPluginCommand(args []string) error {
	if len(args) == 0 {
		return nil
	}

	handler := NewDefaultPluginHandler([]string{PluginPrefix})
	if handler == nil {
		return nil
	}

	if _, _, err := rootCmd.Find(args); err == nil {
		return nil
	}

	var cmdName string
	for _, arg := range args {
		if strings.HasPrefix(arg, "-") {
			continue
		}
		cmdName = arg
		break
	}

	switch cmdName {
	case "", "help", cobra.ShellCompRequestCmd, cobra.ShellCompNoDescRequestCmd:
		return nil
	}

	return HandlePluginCommand(handler, args, 1)
}
