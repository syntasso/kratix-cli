package cmd

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/Masterminds/semver/v3"
	"github.com/spf13/cobra"
)

const (
	skeliftPlugin          = "skelift"
	defaultAPIBaseURL      = "https://api.github.com"
	enterpriseReleasesRepo = "syntasso/enterprise-releases"
	releasesPath           = "/repos/" + enterpriseReleasesRepo + "/releases"
	githubAPIVersion       = "2022-11-28"
	pluginPerm             = 0o755
)

var skeliftBinaries = []string{"kratix-skelift-review", "kratix-skelift-check"}

type githubAsset struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}

type githubRelease struct {
	TagName string        `json:"tag_name"`
	Assets  []githubAsset `json:"assets"`
}

type PluginAddOptions struct {
	Token      string
	TokenStdin bool

	APIBaseURL string
	InstallDir string
	Client     *http.Client
	OS         string
	Arch       string

	In  io.Reader
	Out io.Writer
}

func newPluginAddCommand() *cobra.Command {
	o := &PluginAddOptions{}

	cmd := &cobra.Command{
		Use:   "add PLUGIN_NAME",
		Short: "Install a Kratix CLI plugin",
		Long:  "Install a Kratix CLI plugin. Requires a token with access to the Syntasso enterprise releases.",
		Example: `  # Install the skelift plugin
  # Read the token from stdin
  cat token.txt | kratix plugin add skelift --token-stdin
  
  # Token from CLI flag
  kratix plugin add skelift --token <token>`,
		Args: cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			o.In = cmd.InOrStdin()
			o.Out = cmd.OutOrStdout()
			return o.Run(args[0])
		},
	}

	cmd.Flags().StringVar(&o.Token, "token", "", "Token with access to the Syntasso enterprise releases")
	cmd.Flags().BoolVar(&o.TokenStdin, "token-stdin", false, "Token with access to the Syntasso enterprise releases from stdin")

	return cmd
}

func (o *PluginAddOptions) Run(plugin string) error {
	if plugin != skeliftPlugin {
		return fmt.Errorf("unknown plugin %q: the only supported plugin is %q", plugin, skeliftPlugin)
	}

	token, err := o.token()
	if err != nil {
		return err
	}

	if err := o.applyDefaults(); err != nil {
		return err
	}

	return o.install(token)
}

func (o *PluginAddOptions) applyDefaults() error {
	if o.APIBaseURL == "" {
		o.APIBaseURL = defaultAPIBaseURL
	}
	if o.OS == "" {
		o.OS = runtime.GOOS
	}
	if o.Arch == "" {
		o.Arch = runtime.GOARCH
	}
	if o.Client == nil {
		o.Client = &http.Client{Timeout: 5 * time.Minute}
	}
	if o.InstallDir == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("failed to locate the home directory: %w", err)
		}
		o.InstallDir = filepath.Join(home, ".kratix", "plugins", "bin")
	}
	return nil
}

// The CLI resolves plugins with exec.LookPath, so $PATH is the only place an
// installed plugin will be found.
func (o *PluginAddOptions) onPath() bool {
	for _, dir := range filepath.SplitList(os.Getenv("PATH")) {
		if dir == o.InstallDir {
			return true
		}
	}
	return false
}

func (o *PluginAddOptions) install(token string) error {
	releases, err := o.listReleases(token)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(o.InstallDir, pluginPerm); err != nil {
		return fmt.Errorf("failed to create %s: %w", o.InstallDir, err)
	}

	for _, binary := range skeliftBinaries {
		dest := filepath.Join(o.InstallDir, binary)
		if _, err := os.Stat(dest); err == nil {
			fmt.Fprintf(o.Out, "%s is already installed at %s, skipping\n", binary, dest)
			continue
		}

		asset, tag, err := o.latestAsset(releases, binary)
		if err != nil {
			return err
		}

		if err := o.download(token, asset.URL, dest); err != nil {
			return fmt.Errorf("failed to install %s: %w", binary, err)
		}

		fmt.Fprintf(o.Out, "Installed %s %s to %s\n", binary, strings.TrimPrefix(tag, binary+"-"), o.InstallDir)
	}

	if !o.onPath() {
		fmt.Fprintf(o.Out, "\n%s is not on your PATH. Add it to your shell profile to use the plugins:\n\n    export PATH=\"%s:$PATH\"\n", o.InstallDir, o.InstallDir)
	}

	return nil
}

func (o *PluginAddOptions) listReleases(token string) ([]githubRelease, error) {
	body, err := o.get(token, o.APIBaseURL+releasesPath, "application/vnd.github+json")
	if err != nil {
		return nil, err
	}
	defer body.Close()

	var releases []githubRelease
	if err := json.NewDecoder(body).Decode(&releases); err != nil {
		return nil, fmt.Errorf("failed to parse the enterprise releases response: %w", err)
	}

	return releases, nil
}

func (o *PluginAddOptions) latestAsset(releases []githubRelease, binary string) (githubAsset, string, error) {
	prefix := binary + "-"

	var latest *semver.Version
	var found githubRelease
	for _, release := range releases {
		if !strings.HasPrefix(release.TagName, prefix) {
			continue
		}
		version, err := semver.NewVersion(strings.TrimPrefix(release.TagName, prefix))
		if err != nil {
			continue
		}
		if latest == nil || version.GreaterThan(latest) {
			latest = version
			found = release
		}
	}

	if latest == nil {
		return githubAsset{}, "", fmt.Errorf("no releases found for %s", binary)
	}

	assetName := fmt.Sprintf("%s_%s_%s", binary, o.OS, o.Arch)
	for _, asset := range found.Assets {
		if asset.Name == assetName {
			return asset, found.TagName, nil
		}
	}

	return githubAsset{}, "", fmt.Errorf("no %s build for %s/%s in %s", binary, o.OS, o.Arch, found.TagName)
}

func (o *PluginAddOptions) download(token, url, dest string) error {
	body, err := o.get(token, url, "application/octet-stream")
	if err != nil {
		return err
	}
	defer body.Close()

	// Download to a temporary file and rename, so an interrupted download never
	// leaves a half-written binary on the PATH.
	tmp, err := os.CreateTemp(filepath.Dir(dest), filepath.Base(dest)+".tmp")
	if err != nil {
		return err
	}
	defer os.Remove(tmp.Name())

	if _, err := io.Copy(tmp, body); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	if err := os.Chmod(tmp.Name(), pluginPerm); err != nil {
		return err
	}

	return os.Rename(tmp.Name(), dest)
}

func (o *PluginAddOptions) get(token, url, accept string) (io.ReadCloser, error) {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", accept)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("X-GitHub-Api-Version", githubAPIVersion)

	resp, err := o.Client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to reach the enterprise releases: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden || resp.StatusCode == http.StatusNotFound {
			return nil, fmt.Errorf("token was rejected by the enterprise releases (HTTP %d): check it is valid and has access to %s", resp.StatusCode, enterpriseReleasesRepo)
		}
		return nil, fmt.Errorf("unexpected response from the enterprise releases: HTTP %d", resp.StatusCode)
	}

	return resp.Body, nil
}

func (o *PluginAddOptions) token() (string, error) {
	if o.Token != "" && o.TokenStdin {
		return "", fmt.Errorf("provide the token with either --token or --token-stdin, not both")
	}

	if o.TokenStdin {
		raw, err := io.ReadAll(o.In)
		if err != nil {
			return "", fmt.Errorf("failed to read token from stdin: %w", err)
		}
		o.Token = strings.TrimSpace(string(raw))
	}

	if o.Token == "" {
		return "", fmt.Errorf("no token supplied: use --token or --token-stdin to provide a token")
	}

	return o.Token, nil
}
