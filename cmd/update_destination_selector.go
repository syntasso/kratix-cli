package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/syntasso/kratix/api/v1alpha1"
	"os"
	"path/filepath"
	"sigs.k8s.io/yaml"
	"strings"
)

var updateDestinationSelector = &cobra.Command{
	Use:   "destination-selector KEY=VALUE",
	Short: "Command to update destination selectors",
	Long:  "Command to update destination selectors",
	Example: `  # adds and updates a destination selector
  kratix update destination-selector env=dev
  # removes an existing destination selector
  kratix update destination-selector zone-
`,
	RunE: UpdateSelector,
	Args: cobra.ExactArgs(1),
}

func init() {
	updateCmd.AddCommand(updateDestinationSelector)
	updateDestinationSelector.Flags().StringVarP(&dir, "dir", "d", ".", "Directory to read Promise from")
}

func UpdateSelector(cmd *cobra.Command, args []string) error {
	promise, err := getPromise(filepath.Join(dir, "promise.yaml"))
	if err != nil {
		return fmt.Errorf("failed to find promise.yaml in directory: %v", err)
	}

	if parsed := strings.Split(args[0], "="); len(parsed) == 2 {
		if len(promise.Spec.DestinationSelectors) == 0 {
			promise.Spec.DestinationSelectors = []v1alpha1.PromiseScheduling{{MatchLabels: map[string]string{}}}
		}
		key, value := parsed[0], parsed[1]
		promise.Spec.DestinationSelectors[0].MatchLabels[key] = value
	} else {
		if args[0][len(args[0])-1:] != "-" {
			return fmt.Errorf("invalid destination key: %s", args[0])
		}
		key := strings.TrimRight(args[0], "-")
		if len(promise.Spec.DestinationSelectors) > 0 {
			delete(promise.Spec.DestinationSelectors[0].MatchLabels, key)
		}
	}

	var promiseBytes []byte
	if promiseBytes, err = yaml.Marshal(promise); err != nil {
		return err
	}
	if err = os.WriteFile(filepath.Join(dir, "promise.yaml"), promiseBytes, filePerm); err != nil {
		return err
	}

	fmt.Println("Promise destination selector updated")
	return nil
}
