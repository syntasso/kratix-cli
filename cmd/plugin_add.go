package cmd

import (
	"archive/tar"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
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
	skillPerm              = 0o644

	// The skills artifact is a few hundred kilobytes of markdown; the cap is
	// only here so a malformed archive cannot fill the disk.
	maxSkillsArchiveBytes = 64 << 20
)

type skillsArtifact struct {
	tagPrefix string
	asset     string
}

var skillsArtifacts = []skillsArtifact{
	{tagPrefix: "skelift-skills-", asset: "skelift-skills.tar.gz"},
	{tagPrefix: "kratix-skills-", asset: "kratix-skills.tar.gz"},
}

var skillVersionPattern = regexp.MustCompile(`(?m)^\s*version:\s*"([^"]+)"`)

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

	APIBaseURL      string
	InstallDir      string
	SkillsDir       string
	ClaudeSkillsDir string
	Client          *http.Client
	OS              string
	Arch            string

	In  io.Reader
	Out io.Writer
}

func newPluginAddCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add PLUGIN_NAME",
		Short: "Install a Kratix plugin",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return cmd.Help()
			}
			return fmt.Errorf("unknown plugin %q: the only supported plugin is %q", args[0], skeliftPlugin)
		},
	}

	cmd.AddCommand(newPluginAddSkeliftCommand())

	return cmd
}

func newPluginAddSkeliftCommand() *cobra.Command {
	o := &PluginAddOptions{}

	cmd := &cobra.Command{
		Use:   skeliftPlugin,
		Short: "Install the skelift CLI plugins and skills",
		Long: `Install the skelift CLI plugins and skills.

This adds:

  - CLI plugins to ~/.kratix/plugins/bin, so the kratix CLI can run them. Add
    that directory to your PATH if it is not already there.
  - skills to ~/.kratix/skills, for any agent to use. Point Codex, Kiro or
    Claude Desktop at this directory.
  - a copy of those skills to ~/.claude/skills, where Claude Code picks them up
    automatically. Restart Claude Code to see them.

Anything already installed at those paths is overwritten.

Requires a token with access to the Syntasso enterprise releases.`,
		Example: `  # Read the token from stdin
  cat token.txt | kratix plugin add skelift --token-stdin

  # Token from CLI flag
  kratix plugin add skelift --token <token>`,
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			o.In = cmd.InOrStdin()
			o.Out = cmd.OutOrStdout()
			return o.Run(skeliftPlugin)
		},
	}

	cmd.Flags().StringVar(&o.Token, "token", "", "Token with access to the Syntasso enterprise releases")
	cmd.Flags().BoolVar(&o.TokenStdin, "token-stdin", false, "Read the token from stdin")

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
	if o.SkillsDir == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("failed to locate the home directory: %w", err)
		}
		o.SkillsDir = filepath.Join(home, ".kratix", "skills")
	}
	if o.ClaudeSkillsDir == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("failed to locate the home directory: %w", err)
		}
		o.ClaudeSkillsDir = filepath.Join(home, ".claude", "skills")
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
			fmt.Fprintf(o.Out, "Warning: %s already installed at %s, overwriting\n", binary, o.InstallDir)
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

	var skills []string
	for _, artifact := range skillsArtifacts {
		installed, err := o.installSkills(token, releases, artifact)
		if err != nil {
			return err
		}
		skills = append(skills, installed...)
	}

	if len(skills) > 0 {
		fmt.Fprintf(o.Out, "\nFor other agents (Codex, Kiro, Claude Desktop), point them at %s\n", o.SkillsDir)
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

func latestRelease(releases []githubRelease, prefix string) (githubRelease, *semver.Version, error) {
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
		return githubRelease{}, nil, fmt.Errorf("no releases found matching %s*", prefix)
	}

	return found, latest, nil
}

func (o *PluginAddOptions) latestAsset(releases []githubRelease, binary string) (githubAsset, string, error) {
	found, _, err := latestRelease(releases, binary+"-")
	if err != nil {
		return githubAsset{}, "", err
	}

	assetName := fmt.Sprintf("%s_%s_%s", binary, o.OS, o.Arch)
	for _, asset := range found.Assets {
		if asset.Name == assetName {
			return asset, found.TagName, nil
		}
	}

	return githubAsset{}, "", fmt.Errorf("no %s build for %s/%s in %s", binary, o.OS, o.Arch, found.TagName)
}

func (o *PluginAddOptions) installSkills(token string, releases []githubRelease, artifact skillsArtifact) ([]string, error) {
	release, version, err := latestRelease(releases, artifact.tagPrefix)
	if err != nil {
		return nil, err
	}

	var asset githubAsset
	for _, candidate := range release.Assets {
		if candidate.Name == artifact.asset {
			asset = candidate
			break
		}
	}
	if asset.URL == "" {
		return nil, fmt.Errorf("%s has no %s asset", release.TagName, artifact.asset)
	}

	body, err := o.get(token, asset.URL, "application/octet-stream")
	if err != nil {
		return nil, err
	}
	defer body.Close()

	if err := os.MkdirAll(o.SkillsDir, pluginPerm); err != nil {
		return nil, fmt.Errorf("failed to create %s: %w", o.SkillsDir, err)
	}

	// Staged alongside the skills directory rather than in the system temp dir,
	// so moving each skill into place is a rename within one filesystem.
	staging, err := os.MkdirTemp(filepath.Dir(o.SkillsDir), ".kratix-skills-")
	if err != nil {
		return nil, err
	}
	defer os.RemoveAll(staging)

	if err := extractTarGz(body, staging); err != nil {
		return nil, fmt.Errorf("failed to extract %s: %w", artifact.asset, err)
	}

	entries, err := os.ReadDir(staging)
	if err != nil {
		return nil, err
	}

	var installed []string
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		name := entry.Name()
		dest := filepath.Join(o.SkillsDir, name)

		if _, err := os.Stat(dest); err == nil {
			if installed, ok := skillVersion(dest); ok {
				fmt.Fprintf(o.Out, "Warning: %s %s already installed at %s, overwriting\n", name, installed, o.SkillsDir)
			} else {
				fmt.Fprintf(o.Out, "Warning: %s already installed at %s, overwriting\n", name, o.SkillsDir)
			}
			if err := os.RemoveAll(dest); err != nil {
				return nil, err
			}
		}

		if err := os.Rename(filepath.Join(staging, name), dest); err != nil {
			return nil, err
		}

		fmt.Fprintf(o.Out, "Installed skill %s %s to %s\n", name, version, o.SkillsDir)

		if err := o.copyForClaudeCode(name, dest); err != nil {
			return nil, err
		}

		installed = append(installed, name)
	}

	return installed, nil
}

func (o *PluginAddOptions) copyForClaudeCode(name, source string) error {
	if err := os.MkdirAll(o.ClaudeSkillsDir, pluginPerm); err != nil {
		return fmt.Errorf("failed to create %s: %w", o.ClaudeSkillsDir, err)
	}

	dest := filepath.Join(o.ClaudeSkillsDir, name)
	if _, err := os.Stat(dest); err == nil {
		if installed, ok := skillVersion(dest); ok {
			fmt.Fprintf(o.Out, "Warning: %s %s already installed at %s, overwriting\n", name, installed, o.ClaudeSkillsDir)
		} else {
			fmt.Fprintf(o.Out, "Warning: %s already installed at %s, overwriting\n", name, o.ClaudeSkillsDir)
		}
		if err := os.RemoveAll(dest); err != nil {
			return err
		}
	}

	if err := os.CopyFS(dest, os.DirFS(source)); err != nil {
		return fmt.Errorf("failed to install %s for Claude Code: %w", name, err)
	}

	fmt.Fprintf(o.Out, "  also installed for Claude Code at %s\n", o.ClaudeSkillsDir)
	return nil
}

func skillVersion(dir string) (string, bool) {
	body, err := os.ReadFile(filepath.Join(dir, "SKILL.md"))
	if err != nil {
		return "", false
	}
	if match := skillVersionPattern.FindSubmatch(body); match != nil {
		return string(match[1]), true
	}
	return "", false
}

func extractTarGz(r io.Reader, dest string) error {
	gz, err := gzip.NewReader(r)
	if err != nil {
		return err
	}
	defer gz.Close()

	reader := tar.NewReader(io.LimitReader(gz, maxSkillsArchiveBytes))
	for {
		header, err := reader.Next()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}

		target, err := safeJoin(dest, header.Name)
		if err != nil {
			return err
		}

		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(target, pluginPerm); err != nil {
				return err
			}
		case tar.TypeReg:
			if err := os.MkdirAll(filepath.Dir(target), pluginPerm); err != nil {
				return err
			}
			file, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, skillPerm)
			if err != nil {
				return err
			}
			if _, err := io.Copy(file, reader); err != nil {
				file.Close()
				return err
			}
			if err := file.Close(); err != nil {
				return err
			}
		}
		// Anything else is skipped: the artifact holds only files and
		// directories, and a symlink could point outside the skills directory.
	}
}

func safeJoin(dest, name string) (string, error) {
	target := filepath.Join(dest, name)
	if target != filepath.Clean(dest) && !strings.HasPrefix(target, filepath.Clean(dest)+string(os.PathSeparator)) {
		return "", fmt.Errorf("unsafe path in archive: %s", name)
	}
	return target, nil
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
