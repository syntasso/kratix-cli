package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
)

const PluginPrefix = "kratix"

// PluginHandler is capable of parsing command line arguments and performing
// executable filename lookups to search for valid plugin files, and execute
// found plugins.
type PluginHandler interface {
	Lookup(filename string) (string, bool)
	Execute(executablePath string, cmdArgs, environment []string) error
}

// DefaultPluginHandler implements PluginHandler
type DefaultPluginHandler struct {
	ValidPrefixes []string
}

// NewDefaultPluginHandler instantiates the DefaultPluginHandler with a list of
// given filename prefixes used to identify valid plugin filenames.
func NewDefaultPluginHandler(validPrefixes []string) *DefaultPluginHandler {
	return &DefaultPluginHandler{ValidPrefixes: validPrefixes}
}

// Lookup implements PluginHandler and attempts to locate an executable with the
// provided filename using the configured prefixes.
func (h *DefaultPluginHandler) Lookup(filename string) (string, bool) {
	for _, prefix := range h.ValidPrefixes {
		path, err := exec.LookPath(fmt.Sprintf("%s-%s", prefix, filename))
		if err != nil || len(path) == 0 {
			continue
		}
		return path, true
	}
	return "", false
}

// Command returns an exec.Cmd configured with the provided executable path and
// arguments. It mirrors the helper used by kubectl to retain behaviour on
// Windows where LookPath may resolve executable extensions.
func Command(name string, arg ...string) *exec.Cmd {
	cmd := &exec.Cmd{
		Path: name,
		Args: append([]string{name}, arg...),
	}
	if filepath.Base(name) == name {
		if lp, err := exec.LookPath(name); err == nil && lp != "" {
			cmd.Path = lp
		}
	}
	return cmd
}

// Execute implements PluginHandler and executes the plugin binary with the
// provided arguments.
func (h *DefaultPluginHandler) Execute(executablePath string, cmdArgs, environment []string) error {
	return syscall.Exec(executablePath, append([]string{executablePath}, cmdArgs...), environment)
}

// HandlePluginCommand receives a pluginHandler and command-line arguments and
// attempts to find a plugin executable on the PATH that satisfies the given
// arguments. If a matching plugin is found it is executed and this function does
// not return unless an error occurs.
func HandlePluginCommand(pluginHandler PluginHandler, cmdArgs []string, minArgs int) error {
	var remainingArgs []string
	for _, arg := range cmdArgs {
		if strings.HasPrefix(arg, "-") {
			break
		}
		remainingArgs = append(remainingArgs, strings.ReplaceAll(arg, "-", "_"))
	}

	if len(remainingArgs) == 0 {
		return fmt.Errorf("flags cannot be placed before plugin name: %s", cmdArgs[0])
	}

	var foundBinaryPath string
	for len(remainingArgs) > 0 {
		if path, found := pluginHandler.Lookup(strings.Join(remainingArgs, "-")); found {
			foundBinaryPath = path
			break
		}
		remainingArgs = remainingArgs[:len(remainingArgs)-1]
		if len(remainingArgs) < minArgs {
			break
		}
	}

	if len(foundBinaryPath) == 0 {
		return nil
	}

	if err := pluginHandler.Execute(foundBinaryPath, cmdArgs[len(remainingArgs):], os.Environ()); err != nil {
		return err
	}

	return nil
}
