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
		func(scenarioDir string, expectError bool, expectedMismatchedDeps []checks.DependencyInfoPair) {

			plugin := filepath.Join(testFileDir, scenarioDir, pluginModFileName)
			gloo := filepath.Join(testFileDir, scenarioDir, glooModFileName)

			mismatchedDeps, err := checks.CompareDependencies(plugin, gloo)
			if expectError {
				Expect(err).To(HaveOccurred())
			} else {
				Expect(err).NotTo(HaveOccurred())
				Expect(mismatchedDeps).To(BeEquivalentTo(expectedMismatchedDeps))
			}
		},
		Entry("succeeds if deps are compatible", "success", false, nil),
		Entry("fails if a file is malformed", "malformed", true, nil),
		Entry("fails if dependencies are not compatible", "mismatch", false,
			[]checks.DependencyInfoPair{{
				Message:      "Please pin your dependency to the same version as the Gloo one using a [require] clause",
				MismatchType: checks.Require,
				Plugin: checks.DependencyInfo{
					Name:    "github.com/solo-io/foo",
					Version: "v0.0.0-20180207000608-aaaaaaaaaaaa",
				},
				Gloo: checks.DependencyInfo{
					Name:    "github.com/solo-io/foo",
					Version: "v0.0.0-20180207000608-0eeff89b0690",
				},
			}},
		),
		Entry("succeeds if deps with replacements are compatible", "success_replacements", false, nil),
		Entry("fails if gloo replaces a dep but the plugin does not", "mismatch_replace_1", false,
			[]checks.DependencyInfoPair{{
				Message:      "Please add a [replace] clause matching the Gloo one",
				MismatchType: checks.PluginMissingReplace,
				Plugin: checks.DependencyInfo{
					Name:    "github.com/solo-io/bar",
					Version: "v1.23.3",
				},
				Gloo: checks.DependencyInfo{
					Name:               "github.com/solo-io/bar",
					Version:            "v1.2.3",
					Replacement:        true,
					ReplacementName:    "github.com/solo-io/bar",
					ReplacementVersion: "v1.2.4",
				},
			}},
		),
		Entry("fails if the plugin replaces a dep but Gloo does not", "mismatch_replace_2", false,
			[]checks.DependencyInfoPair{{
				Message:      "Please remove the [replace] clause and pin your dependency to the same version as the Gloo one using a [require] clause",
				MismatchType: checks.PluginExtraReplace,
				Plugin: checks.DependencyInfo{
					Name:               "github.com/solo-io/bar",
					Version:            "v1.2.3",
					Replacement:        true,
					ReplacementName:    "github.com/solo-io/bar",
					ReplacementVersion: "v1.2.4",
				},
				Gloo: checks.DependencyInfo{
					Name:    "github.com/solo-io/bar",
					Version: "v1.2.20",
				},
			}},
		),
		Entry("fails if both the plugin and Gloo replace a dep but the replacements do not match", "mismatch_replace_3", false,
			[]checks.DependencyInfoPair{{
				Message:      "The plugin [replace] clause must match the Gloo one",
				MismatchType: checks.ReplaceMismatch,
				Plugin: checks.DependencyInfo{
					Name:               "github.com/solo-io/bar",
					Version:            "v1.2.3",
					Replacement:        true,
					ReplacementName:    "github.com/solo-io/bar",
					ReplacementVersion: "v1.2.4",
				},
				Gloo: checks.DependencyInfo{
					Name:               "github.com/solo-io/bar",
					Version:            "v1.2.3",
					Replacement:        true,
					ReplacementName:    "github.com/solo-io/bar",
					ReplacementVersion: "v1.2.5",
				},
			}},
		),
	)
})
