package cmd

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

// fakeReleases serves the subset of the GitHub releases API that plugin add
// uses: listing releases, and downloading an asset by id.
func fakeReleases(token string, releases map[string]map[string]string) *httptest.Server {
	var server *httptest.Server
	mux := http.NewServeMux()

	authorised := func(w http.ResponseWriter, r *http.Request) bool {
		if r.Header.Get("Authorization") != "Bearer "+token {
			w.WriteHeader(http.StatusUnauthorized)
			return false
		}
		return true
	}

	mux.HandleFunc("/repos/syntasso/enterprise-releases/releases", func(w http.ResponseWriter, r *http.Request) {
		if !authorised(w, r) {
			return
		}
		var out []map[string]any
		for tag, assets := range releases {
			var assetList []map[string]any
			for name := range assets {
				assetList = append(assetList, map[string]any{
					"name": name,
					"url":  fmt.Sprintf("%s/assets/%s/%s", server.URL, tag, name),
				})
			}
			out = append(out, map[string]any{"tag_name": tag, "assets": assetList})
		}
		w.Header().Set("Content-Type", "application/json")
		Expect(json.NewEncoder(w).Encode(out)).To(Succeed())
	})

	mux.HandleFunc("/assets/", func(w http.ResponseWriter, r *http.Request) {
		if !authorised(w, r) {
			return
		}
		parts := strings.Split(strings.TrimPrefix(r.URL.Path, "/assets/"), "/")
		body, ok := releases[parts[0]][parts[1]]
		if !ok {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		_, err := io.WriteString(w, body)
		Expect(err).NotTo(HaveOccurred())
	})

	server = httptest.NewServer(mux)
	DeferCleanup(server.Close)
	return server
}

// tarball gzips a tar of the given paths, mirroring the skills artifact
// published to enterprise-releases.
func tarball(files map[string]string) string {
	buf := &bytes.Buffer{}
	gz := gzip.NewWriter(buf)
	tw := tar.NewWriter(gz)

	for name, body := range files {
		hdr := &tar.Header{Name: name, Mode: 0o644, Size: int64(len(body)), Typeflag: tar.TypeReg}
		if body == "" && strings.HasSuffix(name, "/") {
			hdr = &tar.Header{Name: name, Mode: 0o755, Typeflag: tar.TypeDir}
		}
		Expect(tw.WriteHeader(hdr)).To(Succeed())
		_, err := tw.Write([]byte(body))
		Expect(err).NotTo(HaveOccurred())
	}

	Expect(tw.Close()).To(Succeed())
	Expect(gz.Close()).To(Succeed())
	return buf.String()
}

var _ = Describe("plugin add", func() {
	const token = "a-token"

	var (
		installDir      string
		skillsDir       string
		claudeSkillsDir string
		out             *strings.Builder
		opts            *PluginAddOptions
	)

	skillFiles := map[string]string{
		"cloud-to-kratix-promise/":                       "",
		"cloud-to-kratix-promise/SKILL.md":               "name: cloud-to-kratix-promise\n  version: \"0.0.2\" # x-release-please-version\n",
		"cloud-to-kratix-promise/examples/shopco.md":     "example",
		"cloud-to-kratix-promise/agent-constraints/x.md": "constraint",
	}

	releases := map[string]map[string]string{
		"kratix-skelift-review-v0.1.0": {"kratix-skelift-review_linux_amd64": "review binary"},
		"kratix-skelift-check-v0.1.0":  {"kratix-skelift-check_linux_amd64": "check binary"},
		"skelift-skills-v0.0.2":        {"skelift-skills.tar.gz": tarball(skillFiles)},
	}

	BeforeEach(func() {
		server := fakeReleases(token, releases)
		installDir = GinkgoT().TempDir()
		skillsDir = GinkgoT().TempDir()
		claudeSkillsDir = GinkgoT().TempDir()
		out = &strings.Builder{}
		opts = &PluginAddOptions{
			Token:           token,
			APIBaseURL:      server.URL,
			InstallDir:      installDir,
			SkillsDir:       skillsDir,
			ClaudeSkillsDir: claudeSkillsDir,
			OS:              "linux",
			Arch:            "amd64",
			Out:             out,
		}
	})

	It("installs the skelift binaries, executable", func() {
		Expect(opts.Run(skeliftPlugin)).To(Succeed())

		for name, contents := range map[string]string{
			"kratix-skelift-review": "review binary",
			"kratix-skelift-check":  "check binary",
		} {
			path := filepath.Join(installDir, name)
			Expect(os.ReadFile(path)).To(BeEquivalentTo(contents))

			info, err := os.Stat(path)
			Expect(err).NotTo(HaveOccurred())
			Expect(info.Mode().Perm()).To(Equal(os.FileMode(0o755)))
		}
	})

	When("a binary is already installed", func() {
		var existing string

		BeforeEach(func() {
			existing = filepath.Join(installDir, "kratix-skelift-review")
			Expect(os.WriteFile(existing, []byte("installed earlier"), 0o755)).To(Succeed())
		})

		It("overwrites it", func() {
			Expect(opts.Run(skeliftPlugin)).To(Succeed())
			Expect(os.ReadFile(existing)).To(BeEquivalentTo("review binary"))
		})

		It("warns before overwriting", func() {
			Expect(opts.Run(skeliftPlugin)).To(Succeed())
			Expect(out.String()).To(ContainSubstring(
				"Warning: kratix-skelift-review already installed at " + installDir + ", overwriting"))
		})
	})

	When("there is no build for the platform", func() {
		BeforeEach(func() {
			opts.OS = "windows"
			opts.Arch = "arm64"
		})

		It("errors naming the platform", func() {
			Expect(opts.Run(skeliftPlugin)).To(MatchError(ContainSubstring("no kratix-skelift-review build for windows/arm64")))
		})
	})

	When("the token is rejected", func() {
		BeforeEach(func() {
			opts.Token = "a-wrong-token"
		})

		It("errors saying the token was rejected", func() {
			Expect(opts.Run(skeliftPlugin)).To(MatchError(ContainSubstring("token was rejected")))
		})
	})

	When("the install dir is not on PATH", func() {
		BeforeEach(func() {
			GinkgoT().Setenv("PATH", "/usr/bin:/bin")
		})

		It("tells the user how to add it", func() {
			Expect(opts.Run(skeliftPlugin)).To(Succeed())
			Expect(out.String()).To(ContainSubstring(`export PATH="` + installDir + `:$PATH"`))
		})
	})

	When("the install dir is already on PATH", func() {
		BeforeEach(func() {
			GinkgoT().Setenv("PATH", "/usr/bin:"+installDir+":/bin")
		})

		It("says nothing about PATH", func() {
			Expect(opts.Run(skeliftPlugin)).To(Succeed())
			Expect(out.String()).NotTo(ContainSubstring("export PATH"))
		})
	})

	Describe("skills", func() {
		It("extracts each skill into the kratix skills directory", func() {
			Expect(opts.Run(skeliftPlugin)).To(Succeed())

			skill := filepath.Join(skillsDir, "cloud-to-kratix-promise")
			Expect(os.ReadFile(filepath.Join(skill, "SKILL.md"))).To(ContainSubstring("cloud-to-kratix-promise"))
			Expect(os.ReadFile(filepath.Join(skill, "examples", "shopco.md"))).To(BeEquivalentTo("example"))
			Expect(os.ReadFile(filepath.Join(skill, "agent-constraints", "x.md"))).To(BeEquivalentTo("constraint"))
		})

		It("also copies each skill into the Claude Code skills directory", func() {
			Expect(opts.Run(skeliftPlugin)).To(Succeed())

			skill := filepath.Join(claudeSkillsDir, "cloud-to-kratix-promise")
			Expect(os.ReadFile(filepath.Join(skill, "SKILL.md"))).To(ContainSubstring("cloud-to-kratix-promise"))
			Expect(os.ReadFile(filepath.Join(skill, "examples", "shopco.md"))).To(BeEquivalentTo("example"))
		})

		It("reports both locations", func() {
			Expect(opts.Run(skeliftPlugin)).To(Succeed())
			Expect(out.String()).To(ContainSubstring("Installed skill cloud-to-kratix-promise 0.0.2 to " + skillsDir))
			Expect(out.String()).To(ContainSubstring("also installed for Claude Code at " + claudeSkillsDir))
		})

		It("points other agents at the kratix skills directory", func() {
			Expect(opts.Run(skeliftPlugin)).To(Succeed())
			Expect(out.String()).To(ContainSubstring("For other agents (Codex, Kiro, Claude Desktop), point them at " + skillsDir))
		})

		When("the skill is already installed", func() {
			BeforeEach(func() {
				skill := filepath.Join(skillsDir, "cloud-to-kratix-promise")
				Expect(os.MkdirAll(skill, 0o755)).To(Succeed())
				Expect(os.WriteFile(filepath.Join(skill, "SKILL.md"),
					[]byte("  version: \"0.0.1\" # x-release-please-version\n"), 0o644)).To(Succeed())
				Expect(os.WriteFile(filepath.Join(skill, "stale.md"), []byte("gone in 0.0.2"), 0o644)).To(Succeed())
			})

			It("warns before overwriting, naming the version it replaces", func() {
				Expect(opts.Run(skeliftPlugin)).To(Succeed())
				Expect(out.String()).To(ContainSubstring(
					"Warning: cloud-to-kratix-promise 0.0.1 already installed at " + skillsDir + ", overwriting"))
			})

			It("does not leave files behind that the new version dropped", func() {
				Expect(opts.Run(skeliftPlugin)).To(Succeed())
				_, err := os.Stat(filepath.Join(skillsDir, "cloud-to-kratix-promise", "stale.md"))
				Expect(os.IsNotExist(err)).To(BeTrue(), "stale.md survived the overwrite")
			})

			It("replaces the Claude Code copy too", func() {
				claudeSkill := filepath.Join(claudeSkillsDir, "cloud-to-kratix-promise")
				Expect(os.MkdirAll(claudeSkill, 0o755)).To(Succeed())
				Expect(os.WriteFile(filepath.Join(claudeSkill, "stale.md"), []byte("old"), 0o644)).To(Succeed())

				Expect(opts.Run(skeliftPlugin)).To(Succeed())

				_, err := os.Stat(filepath.Join(claudeSkill, "stale.md"))
				Expect(os.IsNotExist(err)).To(BeTrue(), "stale.md survived in the Claude Code copy")
				Expect(os.ReadFile(filepath.Join(claudeSkill, "SKILL.md"))).To(ContainSubstring("0.0.2"))
			})
		})

		When("the tarball tries to escape the skills directory", func() {
			BeforeEach(func() {
				escaping := map[string]string{"../escaped.md": "nope"}
				opts.APIBaseURL = fakeReleases(token, map[string]map[string]string{
					"kratix-skelift-review-v0.1.0": {"kratix-skelift-review_linux_amd64": "review binary"},
					"kratix-skelift-check-v0.1.0":  {"kratix-skelift-check_linux_amd64": "check binary"},
					"skelift-skills-v0.0.2":        {"skelift-skills.tar.gz": tarball(escaping)},
				}).URL
			})

			It("errors instead of writing outside it", func() {
				Expect(opts.Run(skeliftPlugin)).To(MatchError(ContainSubstring("unsafe path")))
				Expect(filepath.Join(filepath.Dir(skillsDir), "escaped.md")).NotTo(BeAnExistingFile())
			})
		})

		When("the release has no skills asset", func() {
			BeforeEach(func() {
				opts.APIBaseURL = fakeReleases(token, map[string]map[string]string{
					"kratix-skelift-review-v0.1.0": {"kratix-skelift-review_linux_amd64": "review binary"},
					"kratix-skelift-check-v0.1.0":  {"kratix-skelift-check_linux_amd64": "check binary"},
					"skelift-skills-v0.0.2":        {"something-else.txt": "x"},
				}).URL
			})

			It("errors naming the missing asset", func() {
				Expect(opts.Run(skeliftPlugin)).To(MatchError(ContainSubstring("skelift-skills.tar.gz")))
			})
		})
	})

	Describe("defaults", func() {
		It("installs into the kratix plugin directory under the user's home", func() {
			home, err := os.UserHomeDir()
			Expect(err).NotTo(HaveOccurred())

			o := &PluginAddOptions{}
			Expect(o.applyDefaults()).To(Succeed())
			Expect(o.InstallDir).To(Equal(filepath.Join(home, ".kratix", "plugins", "bin")))
			Expect(o.SkillsDir).To(Equal(filepath.Join(home, ".kratix", "skills")))
			Expect(o.ClaudeSkillsDir).To(Equal(filepath.Join(home, ".claude", "skills")))
		})
	})
})
