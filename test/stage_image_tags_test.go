package integration_test

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	. "github.com/onsi/ginkgo/v2"
)

// Expected-output fixtures write a stage image as
//
//	ghcr.io/syntasso/kratix-cli/<name>:{{stage-tag}}
//
// instead of a literal version. release-please updates the version in cmd/ when
// a stage is released but cannot edit these fixtures (a marker comment would
// stop them matching the generated output), so a literal would fail the tests
// after every stage release. cat() swaps the placeholder for the tag the CLI
// really uses, read from the same source.
var stageTagPlaceholder = regexp.MustCompile(`kratix-cli/([a-z0-9-]+):\{\{stage-tag\}\}`)

func expandStageImageTags(content string) string {
	return stageTagPlaceholder.ReplaceAllStringFunc(content, func(match string) string {
		repository := stageTagPlaceholder.FindStringSubmatch(match)[1]
		return fmt.Sprintf("kratix-cli/%s:%s", repository, stageImageTag(repository))
	})
}

// stageImageTag finds the tag of a stage image the CLI puts into generated promises.
func stageImageTag(repository string) string {
	sources, err := filepath.Glob("../cmd/*.go")
	if err != nil || len(sources) == 0 {
		Fail("could not find the CLI sources in ../cmd")
	}

	literal := regexp.MustCompile(`kratix-cli/` + regexp.QuoteMeta(repository) + `:(v\d+\.\d+\.\d+)`)
	for _, source := range sources {
		if strings.HasSuffix(source, "_test.go") {
			continue
		}
		content, err := os.ReadFile(source)
		if err != nil {
			Fail(fmt.Sprintf("could not read %s: %v", source, err))
		}
		if match := literal.FindSubmatch(content); match != nil {
			return string(match[1])
		}
	}

	// The pulumi image is built from a repository constant and a version constant.
	if repository == "pulumi-generator" {
		content, err := os.ReadFile("../cmd/pulumi_stage_defaults.go")
		if err != nil {
			Fail(fmt.Sprintf("could not read the pulumi defaults: %v", err))
		}
		version := regexp.MustCompile(`pulumiGeneratorImageVersion\s*=\s*"(v[^"]+)"`)
		if match := version.FindSubmatch(content); match != nil {
			return string(match[1])
		}
	}

	Fail(fmt.Sprintf("no tag found in ../cmd for stage image %s", repository))
	return ""
}
