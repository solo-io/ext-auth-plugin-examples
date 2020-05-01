package checks

import (
	"bufio"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	"github.com/pkg/errors"
	"os"
	"path/filepath"
	"strings"
)

const (
	pluginModFileName = "plugin.txt"
	glooModFileName   = "gloo.txt"
)

func parseDependenciesTestFile(filePath string) (map[string]DependencyInfo, error) {
	depFile, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	//noinspection GoUnhandledErrorResult
	defer depFile.Close()

	dependencies := map[string]DependencyInfo{}

	scanner := bufio.NewScanner(depFile)
	skippedFirstLine := false
	for scanner.Scan() {
		line := scanner.Text()

		depInfo := strings.Fields(line)

		// First line is the name of the module the `go list -m all` command was ran for
		if !skippedFirstLine && len(depInfo) == 1 {
			skippedFirstLine = true
			continue
		}

		switch len(depInfo) {
		case 2:
			dependencies[depInfo[0]] = DependencyInfo{
				Name:    depInfo[0],
				Version: depInfo[1],
			}
		case 5:
			dependencies[depInfo[0]] = DependencyInfo{
				Name:               depInfo[0],
				Version:            depInfo[1],
				Replacement:        true,
				ReplacementName:    depInfo[3],
				ReplacementVersion: depInfo[4],
			}
		default:
			return nil, errors.Errorf("malformed dependency: [%s]. "+
				"Expected format 'NAME VERSION' or 'NAME VERSION => REPLACE_NAME REPLACE_VERSION'", line)
		}
	}
	return dependencies, scanner.Err()
}

var _ = Describe("Dependency verification script", func() {

	DescribeTable("can detect incompatible dependency requirements",
		func(scenarioDir string, expectError bool, expectedMismatchedDeps []DependencyInfoPair) {
			var (
				plugin, gloo map[string]DependencyInfo
				err          error
			)
			plugin, err = parseDependenciesTestFile(filepath.Join(testFileDir, scenarioDir, pluginModFileName))
			Expect(err).NotTo(HaveOccurred())
			gloo, err = parseDependenciesTestFile(filepath.Join(testFileDir, scenarioDir, glooModFileName))
			if expectError {
				Expect(err).To(HaveOccurred())
			} else {
				Expect(err).NotTo(HaveOccurred())
			}

			mismatchedDeps := compareDependencies(plugin, gloo)
			Expect(mismatchedDeps).To(BeEquivalentTo(expectedMismatchedDeps))
		},
		Entry("succeeds if deps are compatible", "success", false, nil),
		Entry("fails if a file is malformed", "malformed", true, nil),
		Entry("fails if dependencies are not compatible", "mismatch", false,
			[]DependencyInfoPair{{
				Message:      "Please pin your dependency to the same version as the Gloo one using a [require] clause",
				MismatchType: Require,
				Plugin: DependencyInfo{
					Name:    "github.com/solo-io/foo",
					Version: "v0.0.0-20180207000608-aaaaaaaaaaaa",
				},
				Gloo: DependencyInfo{
					Name:    "github.com/solo-io/foo",
					Version: "v0.0.0-20180207000608-0eeff89b0690",
				},
			}},
		),
		Entry("succeeds if deps with replacements are compatible", "success_replacements", false, nil),
		Entry("fails if gloo replaces a dep but the plugin does not", "mismatch_replace_1", false,
			[]DependencyInfoPair{{
				Message:      "Please add a [replace] clause matching the Gloo one",
				MismatchType: PluginMissingReplace,
				Plugin: DependencyInfo{
					Name:    "github.com/solo-io/bar",
					Version: "v1.23.3",
				},
				Gloo: DependencyInfo{
					Name:               "github.com/solo-io/bar",
					Version:            "v1.2.3",
					Replacement:        true,
					ReplacementName:    "github.com/solo-io/bar",
					ReplacementVersion: "v1.2.4",
				},
			}},
		),
		Entry("fails if the plugin replaces a dep but Gloo does not", "mismatch_replace_2", false,
			[]DependencyInfoPair{{
				Message:      "Please remove the [replace] clause and pin your dependency to the same version as the Gloo one using a [require] clause",
				MismatchType: PluginExtraReplace,
				Plugin: DependencyInfo{
					Name:               "github.com/solo-io/bar",
					Version:            "v1.2.3",
					Replacement:        true,
					ReplacementName:    "github.com/solo-io/bar",
					ReplacementVersion: "v1.2.4",
				},
				Gloo: DependencyInfo{
					Name:    "github.com/solo-io/bar",
					Version: "v1.2.20",
				},
			}},
		),
		Entry("fails if both the plugin and Gloo replace a dep but the replacements do not match", "mismatch_replace_3", false,
			[]DependencyInfoPair{{
				Message:      "The plugin [replace] clause must match the Gloo one",
				MismatchType: ReplaceMismatch,
				Plugin: DependencyInfo{
					Name:               "github.com/solo-io/bar",
					Version:            "v1.2.3",
					Replacement:        true,
					ReplacementName:    "github.com/solo-io/bar",
					ReplacementVersion: "v1.2.4",
				},
				Gloo: DependencyInfo{
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
