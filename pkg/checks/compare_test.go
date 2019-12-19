package checks_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	"github.com/solo-io/ext-auth-plugin-examples/pkg/checks"
	"path/filepath"
)

const (
	testFileDir       = "test"
	pluginModFileName = "plugin.txt"
	glooModFileName   = "gloo.txt"
)

var _ = Describe("Dependency verification script", func() {

	DescribeTable("can detect incompatible dependency requirements",
		func(scenarioDir string, expectedErr *checks.DependencyError) {

			plugin := filepath.Join(testFileDir, scenarioDir, pluginModFileName)
			gloo := filepath.Join(testFileDir, scenarioDir, glooModFileName)

			report, err := checks.CompareDependencies(plugin, gloo)
			Expect(err).NotTo(HaveOccurred())

			if expectedErr == nil {
				Expect(report).To(HaveLen(0))
			} else {
				Expect(report).To(HaveLen(1))
				actualErr := report.GetEntry(expectedErr.Module)
				Expect(actualErr).NotTo(BeNil())
				Expect(actualErr.Kind).To(Equal(expectedErr.Kind))
			}
		},
		Entry("succeeds if deps are compatible", "success_1", nil),
		Entry("succeeds if deps are compatible (with replace)", "success_2", nil),
		Entry("fails if there is a [require] mismatch",
			"require_mismatch",
			&checks.DependencyError{
				Kind:   checks.GlooRequireVersionMismatch,
				Module: "github.com/bar/bar",
			}),
		Entry("fails if gloo replaces a module required by the plugin (different name)",
			"gloo_replaces_req_1",
			&checks.DependencyError{
				Kind:   checks.GlooReplaceNameMismatch,
				Module: "github.com/bar/bar",
			}),
		Entry("fails if gloo replaces a module required by the plugin (different version)",
			"gloo_replaces_req_2",
			&checks.DependencyError{
				Kind:   checks.GlooReplaceVersionMismatch,
				Module: "github.com/bar/bar",
			}),
		Entry("fails if gloo replaces a module and the plugin does not",
			"plugin_missing_replace_1",
			&checks.DependencyError{
				Kind:   checks.PluginMissingReplace,
				Module: "github.com/bar/bar",
			}),
		Entry("fails if gloo and the plugin replace a module in different ways",
			"plugin_missing_replace_2",
			&checks.DependencyError{
				Kind:   checks.PluginReplaceMismatch,
				Module: "github.com/bar/bar",
			}),
	)
})
