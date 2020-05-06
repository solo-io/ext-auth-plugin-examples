package checks

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	"path/filepath"
)

const (
	testFileDir              = "test"
	pluginModuleFileName     = "plugin.mod"
	glooDependenciesFileName = "gloo.txt"
)

var (
	pluginModuleName = "github.com/solo-io/ext-auth-plugin-examples"
	glooModuleName   = "github.com/solo-io/solo-projects"
	moduleVersion    = "1.14"
)

var _ = Describe("parseModule script", func() {

	DescribeTable("can parse module file",
		func(scenarioDir string, expectError bool, expectedModuleInfo *ModuleInfo) {

			plugin := filepath.Join(testFileDir, scenarioDir, pluginModuleFileName)

			moduleInfo, err := parseModuleFile(plugin)
			if expectError {
				Expect(err).To(HaveOccurred())
			} else {
				Expect(err).NotTo(HaveOccurred())
				Expect(moduleInfo).To(BeEquivalentTo(expectedModuleInfo))
			}
		},
		Entry("succeeds if file is welformed", "success_parse", false,
			&ModuleInfo{Name: pluginModuleName, Version: moduleVersion,
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
var _ = Describe("parseDependencies script", func() {

	DescribeTable("can parse dependencies file",
		func(scenarioDir string, expectError bool, expectedModuleInfo *ModuleInfo) {

			gloo := filepath.Join(testFileDir, scenarioDir, glooDependenciesFileName)

			moduleInfo, err := parseDependenciesFile(gloo)
			if expectError {
				Expect(err).To(HaveOccurred())
			} else {
				Expect(err).NotTo(HaveOccurred())
				Expect(moduleInfo).To(BeEquivalentTo(expectedModuleInfo))
			}
		},
		Entry("succeeds if file has no replacements", "success", false,
			&ModuleInfo{Name: glooModuleName,
				Require: map[string]string{
					"github.com/solo-io/bar": "github.com/solo-io/bar v1.2.3",
					"github.com/solo-io/foo": "github.com/solo-io/foo v0.0.0-20180207000608-0eeff89b0690",
				}},
		),
		Entry("succeeds if file has replacements", "success_replacements", false,
			&ModuleInfo{Name: glooModuleName,
				Replace: map[string]string{
					"github.com/solo-io/bar": "github.com/solo-io/bar v1.2.3 => github.com/solo-io/bar v1.2.4",
				},
				Require: map[string]string{
					"github.com/solo-io/foo": "github.com/solo-io/foo v0.0.0-20180207000608-0eeff89b0690",
				}},
		),
		Entry("fails if a file is malformed", "malformed", true, nil),
	)
})
var _ = Describe("MergeModuleFiles script", func() {

	DescribeTable("After merging plugin and gloo dependencies files",
		func(scenarioDir string, expectError bool, expectedModuleInfo *ModuleInfo) {

			plugin := filepath.Join(testFileDir, scenarioDir, pluginModuleFileName)
			gloo := filepath.Join(testFileDir, scenarioDir, glooDependenciesFileName)

			moduleInfo, _, err := MergeModuleFiles(plugin, gloo)
			if expectError {
				Expect(err).To(HaveOccurred())
			} else {
				Expect(err).NotTo(HaveOccurred())
				Expect(moduleInfo).To(BeEquivalentTo(expectedModuleInfo))
			}
		},
		Entry("Gloo require version takes precedence if both the plugin and Gloo require a dep and the version do not match", "success", false,
			&ModuleInfo{Name: pluginModuleName, Version: moduleVersion,
				Require: map[string]string{
					"github.com/solo-io/bar": "github.com/solo-io/bar v1.2.3",
					"github.com/solo-io/baz": "github.com/solo-io/baz v1.2.5",
					"github.com/solo-io/foo": "github.com/solo-io/foo v0.0.0-20180207000608-0eeff89b0690",
				}},
		),
		Entry("Gloo replacement is added for the require dep with matching version", "mismatch_replace_1", false,
			&ModuleInfo{Name: pluginModuleName, Version: moduleVersion,
				Replace: map[string]string{
					"github.com/solo-io/bar": "github.com/solo-io/bar v1.2.3 => github.com/solo-io/bar v1.2.4",
				},
				Require: map[string]string{
					"github.com/solo-io/foo": "github.com/solo-io/foo v0.0.0-20180207000608-0eeff89b0690",
					"github.com/solo-io/bar": "github.com/solo-io/bar v1.23.3",
				}},
		),
		Entry("The plugin replaces a dep which has a different version of the require dep", "mismatch_replace_2", false,
			&ModuleInfo{Name: pluginModuleName, Version: moduleVersion,
				Replace: map[string]string{
					"github.com/solo-io/bar": "github.com/solo-io/bar v1.2.3 => github.com/solo-io/bar v1.2.4",
				},
				Require: map[string]string{
					"github.com/solo-io/foo": "github.com/solo-io/foo v0.0.0-20180207000608-0eeff89b0690",
					"github.com/solo-io/bar": "github.com/solo-io/bar v1.2.20",
				}},
		),
		Entry("Gloo replacement takes precedence if both the plugin and Gloo replace a dep and the replacements do not match", "mismatch_replace_3", false,
			&ModuleInfo{Name: pluginModuleName, Version: moduleVersion,
				Replace: map[string]string{
					"github.com/solo-io/bar": "github.com/solo-io/bar v1.2.3 => github.com/solo-io/bar v1.2.5",
				},
				Require: map[string]string{
					"github.com/solo-io/foo": "github.com/solo-io/foo v0.0.0-20180207000608-0eeff89b0690",
				}},
		),
	)
})
