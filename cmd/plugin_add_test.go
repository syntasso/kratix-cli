package cmd

import (
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

var _ = Describe("plugin add", func() {
	const token = "a-token"

	var (
		installDir string
		out        *strings.Builder
		opts       *PluginAddOptions
	)

	releases := map[string]map[string]string{
		"kratix-skelift-review-v0.1.0": {"kratix-skelift-review_linux_amd64": "review binary"},
		"kratix-skelift-check-v0.1.0":  {"kratix-skelift-check_linux_amd64": "check binary"},
	}

	BeforeEach(func() {
		server := fakeReleases(token, releases)
		installDir = GinkgoT().TempDir()
		out = &strings.Builder{}
		opts = &PluginAddOptions{
			Token:      token,
			APIBaseURL: server.URL,
			InstallDir: installDir,
			OS:         "linux",
			Arch:       "amd64",
			Out:        out,
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

		It("leaves it alone and says so", func() {
			Expect(opts.Run(skeliftPlugin)).To(Succeed())

			Expect(os.ReadFile(existing)).To(BeEquivalentTo("installed earlier"))
			Expect(out.String()).To(ContainSubstring("kratix-skelift-review is already installed"))
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

	Describe("defaults", func() {
		It("installs into the kratix plugin directory under the user's home", func() {
			home, err := os.UserHomeDir()
			Expect(err).NotTo(HaveOccurred())

			o := &PluginAddOptions{}
			Expect(o.applyDefaults()).To(Succeed())
			Expect(o.InstallDir).To(Equal(filepath.Join(home, ".kratix", "plugins", "bin")))
		})
	})
})
