package checks_test

import (
	"path/filepath"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	"github.com/solo-io/ext-auth-plugin-examples/pkg/checks"
)

const (
	testFileDir              = "test"
	pluginModuleFileName     = "plugin.mod"
	glooDependenciesFileName = "gloo.txt"
)

var (
	pluginModuleName = "github.com/solo-io/ext-auth-plugin-examples"
	moduleVersion    = "1.14"
)

var _ = Describe("parseModule script", func() {

	DescribeTable("can parse module file",
		func(scenarioDir string, expectError bool, expectedModuleInfo *checks.ModuleInfo) {

			plugin := filepath.Join(testFileDir, scenarioDir, pluginModuleFileName)

			moduleInfo, err := checks.ParseModuleFile(plugin)
			if expectError {
				Expect(err).To(HaveOccurred())
			} else {
				Expect(err).NotTo(HaveOccurred())
				Expect(moduleInfo).To(BeEquivalentTo(expectedModuleInfo))
			}
		},
		Entry("succeeds if file is well-formed", "success_parse", false,
			&checks.ModuleInfo{Name: pluginModuleName, Version: moduleVersion,
				Replace: map[string]string{
					"github.com/solo-io/bar": "github.com/solo-io/bar v1.2.3 => github.com/solo-io/bar v1.2.4",
					"github.com/solo-io/foo": "github.com/solo-io/foo => github.com/solo-io/bar v1.2.4",
					"github.com/solo-io/baz": "github.com/solo-io/baz v1.2.5 => github.com/solo-io/barfoo v1.2.4",
				},
				Require: map[string]string{
					"github.com/solo-io/bar": "github.com/solo-io/bar v1.2.3 //indirect",
					"github.com/solo-io/baz": "github.com/solo-io/baz v1.2.5",
					"github.com/solo-io/foo": "github.com/solo-io/foo v0.0.0-20180207000608-0eeff89b0690",
				}},
		),
		Entry("fails if a file is malformed", "malformed", true, nil),
	)
})
var _ = Describe("ParseDependencies", func() {

	DescribeTable("can parse dependencies file",
		func(scenarioDir string, expectError bool, expectedDependencyInfo map[string]checks.DependencyInfo) {

			gloo := filepath.Join(testFileDir, scenarioDir, glooDependenciesFileName)

			dependencyInfo, err := checks.ParseDependenciesFile(gloo)
			if expectError {
				Expect(err).To(HaveOccurred())
			} else {
				Expect(err).NotTo(HaveOccurred())
				Expect(dependencyInfo).To(BeEquivalentTo(expectedDependencyInfo))
			}
		},
		Entry("succeeds if file has no replacements", "success", false,
			map[string]checks.DependencyInfo{
				"github.com/solo-io/bar": {
					Name:               "github.com/solo-io/bar",
					Version:            "v1.2.3",
					Replacement:        false,
					ReplacementName:    "",
					ReplacementVersion: "",
				},
				"github.com/solo-io/foo": {
					Name: "github.com/solo-io/foo",
					Version: "v0.0.0-20180207000608-0eeff89b0690",
					Replacement: false,
					ReplacementName: "",
					ReplacementVersion: "",
				},
			},
		),
		Entry("succeeds if file has replacements", "success_replacements", false,
			map[string]checks.DependencyInfo{
				"github.com/solo-io/bar": {
					Name:               "github.com/solo-io/bar",
					Version:            "v1.2.3",
					Replacement:        true,
					ReplacementName:    "github.com/solo-io/bar",
					ReplacementVersion: "v1.2.4",
				},
				"github.com/solo-io/foo": {
					Name: "github.com/solo-io/foo",
					Version: "v0.0.0-20180207000608-0eeff89b0690",
					Replacement: false,
					ReplacementName: "",
					ReplacementVersion: "",
				},
			},
		),
		Entry("fails if a file is malformed", "malformed", true, nil),
	)
})
var _ = PDescribe("MergeModuleFiles", func() {

	DescribeTable("After merging plugin and gloo dependencies files",
		func(scenarioDir string, expectError bool, expectedModuleInfo *checks.ModuleInfo) {

			plugin := filepath.Join(testFileDir, scenarioDir, pluginModuleFileName)
			gloo := filepath.Join(testFileDir, scenarioDir, glooDependenciesFileName)

			moduleInfo, err := checks.MergeModuleFiles(plugin, gloo)
			if expectError {
				Expect(err).To(HaveOccurred())
			} else {
				Expect(err).NotTo(HaveOccurred())
				Expect(moduleInfo).To(BeEquivalentTo(expectedModuleInfo))
			}
		},
		Entry("Gloo require version takes precedence if both the plugin and Gloo require a dep and the version do not match", "success", false,
			&checks.ModuleInfo{Name: pluginModuleName, Version: moduleVersion,
				Require: map[string]string{
					"github.com/solo-io/baz": "github.com/solo-io/baz v1.2.5",
					"github.com/solo-io/foo": "github.com/solo-io/foo v0.0.0-20180207000608-0eeff89b0690",
				},
				Replace: map[string]string{
					"github.com/solo-io/baz": "github.com/solo-io/baz v1.2.5",
					"github.com/solo-io/foo": "github.com/solo-io/foo v0.0.0-20180207000608-0eeff89b0690",
				},
			},
		),
		Entry("Gloo replacement is added for the require dep with matching version", "mismatch_replace_1", false,
			&checks.ModuleInfo{Name: pluginModuleName, Version: moduleVersion,
				Replace: map[string]string{
					"github.com/solo-io/bar": "github.com/solo-io/bar v1.2.3 => github.com/solo-io/bar v1.2.4",
				},
				Require: map[string]string{
					"github.com/solo-io/foo": "github.com/solo-io/foo v0.0.0-20180207000608-0eeff89b0690",
					"github.com/solo-io/bar": "github.com/solo-io/bar v1.23.3",
				}},
		),
		Entry("Gloo require name and version takes precedence if the plugin replaces a require dep", "mismatch_replace_2", false,
			&checks.ModuleInfo{Name: pluginModuleName, Version: moduleVersion,
				Require: map[string]string{
					"github.com/solo-io/foo": "github.com/solo-io/foo v0.0.0-20180207000608-0eeff89b0690",
					"github.com/solo-io/bar": "github.com/solo-io/bar v1.2.20",
				}},
		),
		Entry("Gloo replacement takes precedence if both the plugin and Gloo replace a dep and the replacements version do not match", "mismatch_replace_3", false,
			&checks.ModuleInfo{Name: pluginModuleName, Version: moduleVersion,
				Replace: map[string]string{
					"github.com/solo-io/bar": "github.com/solo-io/bar v1.2.3 => github.com/solo-io/bar v1.2.5",
				},
				Require: map[string]string{
					"github.com/solo-io/foo": "github.com/solo-io/foo v0.0.0-20180207000608-0eeff89b0690",
				}},
		),
		Entry("Plugin replacement takes precedence if both the plugin replace a dep with no version", "mismatch_replace_4", false,
			&checks.ModuleInfo{Name: pluginModuleName, Version: moduleVersion,
				Replace: map[string]string{
					"github.com/solo-io/bar": "github.com/solo-io/bar => github.com/solo-io/baz v1.20.4",
				},
				Require: map[string]string{
					"github.com/solo-io/foo": "github.com/solo-io/foo v0.0.0-20180207000608-0eeff89b0690",
				}},
		),
	)
})
