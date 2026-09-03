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

func tarball(files map[string]string) string {
	buf := &bytes.Buffer{}
	gz := gzip.NewWriter(buf)
	tw := tar.NewWriter(gz)

	for name, body := range files {
		hdr := &tar.Header{Name: name, Mode: 0o644, Size: int64(len(body)), Typeflag: tar.TypeReg}
		if body == "" && strings.HasSuffix(name, "/") {
			hdr = &tar.Header{Name: name, Mode: 0o755, Typeflag: tar.TypeDir}
		}
		ExpectWithOffset(1, tw.WriteHeader(hdr)).To(Succeed())
		_, err := tw.Write([]byte(body))
		ExpectWithOffset(1, err).NotTo(HaveOccurred())
	}

	ExpectWithOffset(1, tw.Close()).To(Succeed())
	ExpectWithOffset(1, gz.Close()).To(Succeed())
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

	kratixSkillFiles := map[string]string{
		"kratix-build-promise/":                   "",
		"kratix-build-promise/SKILL.md":           "name: kratix-build-promise\n  version: \"0.4.0\" # x-release-please-version\n",
		"kratix-build-promise/rules/init/helm.md": "helm",
		"kratix-consume-promise/":                 "",
		"kratix-consume-promise/SKILL.md":         "name: kratix-consume-promise\n  version: \"0.4.0\" # x-release-please-version\n",
		"kratix-consume-promise/rules/kubectl.md": "kubectl",
	}

	releases := map[string]map[string]string{
		"kratix-skelift-review-v0.1.0": {"kratix-skelift-review_linux_amd64": "review binary"},
		"kratix-skelift-check-v0.1.0":  {"kratix-skelift-check_linux_amd64": "check binary"},
		"skelift-skills-v0.0.2":        {"skelift-skills.tar.gz": tarball(skillFiles)},
		"kratix-skills-v0.4.0":         {"kratix-skills.tar.gz": tarball(kratixSkillFiles)},
		"another-release-a":            {"abc": "def"},
		"another-release-b":            {"abc": "def"},
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

	It("installs the skelift binaries", func() {
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
			opts.Token = "an-invalid-token"
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

		It("says nothing about exporting PATH", func() {
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

		It("installs the skills from every artifact", func() {
			Expect(opts.Run(skeliftPlugin)).To(Succeed())

			for skill, version := range map[string]string{
				"cloud-to-kratix-promise": "0.0.2",
				"kratix-build-promise":    "0.4.0",
				"kratix-consume-promise":  "0.4.0",
			} {
				Expect(os.ReadFile(filepath.Join(skillsDir, skill, "SKILL.md"))).To(ContainSubstring(skill))
				Expect(os.ReadFile(filepath.Join(claudeSkillsDir, skill, "SKILL.md"))).To(ContainSubstring(skill))
				Expect(out.String()).To(ContainSubstring("Installed skill " + skill + " " + version))
			}

			Expect(os.ReadFile(filepath.Join(skillsDir, "kratix-build-promise", "rules", "init", "helm.md"))).To(BeEquivalentTo("helm"))
		})

		It("points other agents at the kratix skills directory once, not per artifact", func() {
			Expect(opts.Run(skeliftPlugin)).To(Succeed())
			Expect(strings.Count(out.String(), "For other agents")).To(Equal(1))
		})

		When("one of the skills artifacts has no release", func() {
			BeforeEach(func() {
				opts.APIBaseURL = fakeReleases(token, map[string]map[string]string{
					"kratix-skelift-review-v0.1.0": {"kratix-skelift-review_linux_amd64": "review binary"},
					"kratix-skelift-check-v0.1.0":  {"kratix-skelift-check_linux_amd64": "check binary"},
					"skelift-skills-v0.0.2":        {"skelift-skills.tar.gz": tarball(skillFiles)},
				}).URL
			})

			It("fails rather than installing only some of the skills", func() {
				Expect(opts.Run(skeliftPlugin)).To(MatchError(ContainSubstring("kratix-skills-")))
			})
		})

		It("copies each skill into the Claude Code skills directory", func() {
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

			It("replaces the Claude Code copy", func() {
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

			It("errors with the missing asset", func() {
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
