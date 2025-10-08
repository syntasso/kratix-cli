package cmd

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

func init() {
	rootCmd.AddCommand(newPluginCommand())
}

func newPluginCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "plugin",
		Short: "Provides utilities for interacting with plugins",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}

	cmd.AddCommand(newPluginListCommand())
	return cmd
}

type PluginListOptions struct {
	Verifier    PathVerifier
	PluginPaths []string

	Out    io.Writer
	ErrOut io.Writer
}

func newPluginListCommand() *cobra.Command {
	o := &PluginListOptions{}

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List all visible plugin executables on a user's PATH",
		RunE: func(cmd *cobra.Command, args []string) error {
			o.Out = cmd.OutOrStdout()
			o.ErrOut = cmd.ErrOrStderr()
			if err := o.Complete(cmd); err != nil {
				return err
			}
			return o.Run()
		},
	}

	return cmd
}

func (o *PluginListOptions) Complete(cmd *cobra.Command) error {
	o.Verifier = &CommandOverrideVerifier{root: cmd.Root(), seenPlugins: map[string]string{}}
	o.PluginPaths = filepath.SplitList(os.Getenv("PATH"))
	if o.Out == nil {
		o.Out = cmd.OutOrStdout()
	}
	if o.ErrOut == nil {
		o.ErrOut = cmd.ErrOrStderr()
	}
	return nil
}

func (o *PluginListOptions) Run() error {
	plugins, pluginErrors := o.ListPlugins()

	if len(plugins) > 0 {
		fmt.Fprintf(o.Out, "The following compatible plugins are available:\n\n")
	} else {
		pluginErrors = append(pluginErrors, fmt.Errorf("error: unable to find any kratix plugins in your PATH"))
	}

	pluginWarnings := 0
	for _, pluginPath := range plugins {
		fmt.Fprintf(o.Out, "%s\n", pluginPath)
		if errs := o.Verifier.Verify(pluginPath); len(errs) != 0 {
			for _, err := range errs {
				fmt.Fprintf(o.ErrOut, "  - %s\n", err)
				pluginWarnings++
			}
		}
	}

	if pluginWarnings > 0 {
		if pluginWarnings == 1 {
			pluginErrors = append(pluginErrors, fmt.Errorf("error: one plugin warning was found"))
		} else {
			pluginErrors = append(pluginErrors, fmt.Errorf("error: %d plugin warnings were found", pluginWarnings))
		}
	}

	if len(pluginErrors) > 0 {
		buf := bytes.NewBuffer(nil)
		for _, e := range pluginErrors {
			fmt.Fprintln(buf, e)
		}
		return fmt.Errorf("%s", buf.String())
	}

	return nil
}

func (o *PluginListOptions) ListPlugins() ([]string, []error) {
	var plugins []string
	var errs []error

	for _, dir := range uniquePathsList(o.PluginPaths) {
		if strings.TrimSpace(dir) == "" {
			continue
		}

		files, err := os.ReadDir(dir)
		if err != nil {
			var pathErr *os.PathError
			if errors.As(err, &pathErr) {
				fmt.Fprintf(o.ErrOut, "Unable to read directory %q from your PATH: %v. Skipping...\n", dir, err)
				continue
			}

			errs = append(errs, fmt.Errorf("error: unable to read directory %q in your PATH: %v", dir, err))
			continue
		}

		for _, f := range files {
			if f.IsDir() {
				continue
			}
			if !hasValidPrefix(f.Name()) {
				continue
			}

			fullPath := filepath.Join(dir, f.Name())
			if ok, err := isExecutable(fullPath); err != nil {
				errs = append(errs, fmt.Errorf("error: unable to identify %s as an executable file: %v", fullPath, err))
				continue
			} else if !ok {
				fmt.Fprintf(o.ErrOut, "Skipping plugin %q as it is not executable\n", fullPath)
				continue
			}

			plugins = append(plugins, fullPath)
		}
	}

	return plugins, errs
}

type PathVerifier interface {
	Verify(path string) []error
}

type CommandOverrideVerifier struct {
	root        *cobra.Command
	seenPlugins map[string]string
}

func (v *CommandOverrideVerifier) Verify(path string) []error {
	if v.root == nil {
		return []error{fmt.Errorf("unable to verify path with nil root")}
	}

	binName := filepath.Base(path)

	cmdPath := strings.Split(binName, "-")
	if len(cmdPath) > 1 {
		cmdPath = cmdPath[1:]
	}

	var errs []error

	if existingPath, ok := v.seenPlugins[binName]; ok {
		errs = append(errs, fmt.Errorf("warning: %s is overshadowed by a similarly named plugin: %s", path, existingPath))
	} else {
		v.seenPlugins[binName] = path
	}

	if cmd, _, err := v.root.Find(cmdPath); err == nil {
		errs = append(errs, fmt.Errorf("warning: %s overwrites existing command: %q", binName, cmd.CommandPath()))
	}

	if ok, err := isExecutable(path); err == nil && !ok {
		errs = append(errs, fmt.Errorf("warning: %s identified as a kratix plugin, but it is not executable", path))
	} else if err != nil {
		errs = append(errs, fmt.Errorf("error: unable to identify %s as an executable file: %v", path, err))
	}

	return errs
}

func isExecutable(fullPath string) (bool, error) {
	info, err := os.Stat(fullPath)
	if err != nil {
		return false, err
	}
	
	if m := info.Mode(); !m.IsDir() && m&0o111 != 0 {
		return true, nil
	}

	return false, nil
}

func uniquePathsList(paths []string) []string {
	seen := map[string]bool{}
	var newPaths []string
	for _, p := range paths {
		if seen[p] {
			continue
		}
		seen[p] = true
		newPaths = append(newPaths, p)
	}
	return newPaths
}

func hasValidPrefix(filename string) bool {
	if strings.HasPrefix(filename, PluginPrefix+"-") {
		return true
	}
	return false
}
