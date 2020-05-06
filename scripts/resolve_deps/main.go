package main

import (
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"github.com/solo-io/ext-auth-plugin-examples/pkg/checks"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strconv"
)

const (
	errorReportFile     = "mismatched_dependencies.json"
	suggestionsFileName = "suggestions"
	tempDirName         = "tmp"
)

func main() {
	if len(os.Args) != 4 {
		fmt.Printf("Must provide 2 arguments: \n\t- Plugin go.mod file path\n\t- Glooe dependencies file path\n\t- merge attempts plugin go.mod file\n")
		os.Exit(1)
	}

	pluginsModuleFilePath := os.Args[1]
	glooDependenciesFilePath := os.Args[2]
	var (
		mergeAttempt    int
		nonMatchingDeps []checks.DependencyInfoPair
		err             error
	)
	if mergeAttempt, err = strconv.Atoi(os.Args[3]); err != nil {
		fmt.Printf("Provided 3th arguments is not a number\n")
		os.Exit(1)
	}

	if nonMatchingDeps, err = resolveDependencies(pluginsModuleFilePath, glooDependenciesFilePath, mergeAttempt); err != nil {
		fmt.Printf("Failed to resolve dependencies: %s/n", err.Error())
		os.Exit(1)
	}

	if len(nonMatchingDeps) == 0 {
		fmt.Println("All shared dependencies match")
		os.Exit(0)
	}
	fmt.Printf("Plugin and Gloo Enterprise dependencies do not match after %d merge attempts!\n", mergeAttempt)

	// 1. Write the report to stdout
	reportBytes, err := json.MarshalIndent(nonMatchingDeps, "", "  ")
	if err != nil {
		fmt.Printf("Failed to marshall error report to JSON: %s/n", err.Error())
		os.Exit(1)
	}
	fmt.Println(string(reportBytes))

	// 2. Write the report to a file
	fmt.Printf("Writing error report file [%s] after %d merge attempts\n", errorReportFile, mergeAttempt)
	if err := ioutil.WriteFile(errorReportFile, reportBytes, 0644); err != nil {
		fmt.Printf("Failed to write error report file: %s/n", err.Error())
	}

	// 3. Create a file with suggested changes to go.mod
	if err := createSuggestionsFile(nonMatchingDeps); err != nil {
		fmt.Printf("Failed to create suggestions file: %s/n", err.Error())
	}
	os.Exit(1)
}

func resolveDependencies(moduleFilePath, glooDependenciesFilePath string, mergeAttempt int) ([]checks.DependencyInfoPair, error) {
	var (
		nonMatchingDeps []checks.DependencyInfoPair
		mergedModule    *checks.ModuleInfo
		err             error
	)
	suggestionModuleFileName := moduleFilePath
	for i := 1; mergeAttempt > 0 && i <= mergeAttempt; i++ {
		if mergedModule, nonMatchingDeps, err = checks.MergeModuleFiles(suggestionModuleFileName, glooDependenciesFilePath); err != nil {
			return nil, errors.Wrapf(err, "failed to merge dependencies")
		}

		if len(nonMatchingDeps) == 0 {
			if suggestionModuleFileName != moduleFilePath {
				if err = os.Rename(suggestionModuleFileName, moduleFilePath); err != nil {
					return nil, errors.Wrapf(err, "failed to rename temp suggestions module '%s' to current '%s' file\n", suggestionModuleFileName, moduleFilePath)
				}
			}

			return nonMatchingDeps, nil
		}
		fmt.Println("Plugin and Gloo Enterprise dependencies do not match!")
		if i < mergeAttempt {
			fmt.Printf("Merging dependencies and start comparing again (attempt: %d)\n", i)
		}

		suggestionFilesDir := filepath.Dir((filepath.Join(tempDirName, moduleFilePath)))
		if err := os.MkdirAll(suggestionFilesDir, os.ModePerm); err != nil {
			return nil, errors.Wrapf(err, "failed to create temp suggestions dir '%s' file\n", suggestionFilesDir)
		}

		suggestionModuleFileName = fmt.Sprintf("%s/%s-%d", suggestionFilesDir, moduleFilePath, i)
		if err = createPluginModuleFile(suggestionModuleFileName, mergedModule); err != nil {
			return nil, errors.Wrapf(err, "failed to write new merged '%s' file\n", moduleFilePath)
		}
	}
	return nonMatchingDeps, err
}

func createSuggestionsFile(nonMatchingDeps []checks.DependencyInfoPair) error {
	suggestionsFile, err := os.Create(suggestionsFileName)
	if err != nil {
		return err
	}
	//noinspection GoUnhandledErrorResult
	defer suggestionsFile.Close()

	fmt.Printf("Writing suggestions file [%s], please use its content to update your go.mod file\n", suggestionsFileName)
	suggestionMap := map[checks.MismatchType][]string{}
	for _, pair := range nonMatchingDeps {
		if pair.MismatchType == checks.Require {
			suggestionMap[checks.Require] = append(suggestionMap[checks.Require],
				fmt.Sprintf("%s %s", pair.Gloo.Name, pair.Gloo.Version))
		} else if pair.MismatchType == checks.PluginMissingReplace || pair.MismatchType == checks.ReplaceMismatch {
			suggestionMap[checks.ReplaceMismatch] = append(suggestionMap[checks.ReplaceMismatch],
				fmt.Sprintf("%s %s => %s %s", pair.Gloo.Name, pair.Gloo.Version, pair.Gloo.ReplacementName, pair.Gloo.ReplacementVersion))
		}
	}

	// Print out the suggested changes for the `require` section of the go.mod file
	if requires, ok := suggestionMap[checks.Require]; ok && len(requires) > 0 {
		_, _ = fmt.Fprintln(suggestionsFile, `require (
	// Add the following entries to the 'require' section of your go.mod file:`)
		for _, r := range requires {
			_, _ = fmt.Fprintf(suggestionsFile, "\t%s\n", r)
		}
		_, _ = fmt.Fprintln(suggestionsFile, ")")
	}

	// Print out the suggested changes for the `replace` section of the go.mod file

	if replacements, ok := suggestionMap[checks.ReplaceMismatch]; ok && len(replacements) > 0 {
		_, _ = fmt.Fprintln(suggestionsFile, `replace (
	// Add the following entries to the 'replace' section of your go.mod file:`)
		for _, r := range replacements {
			_, _ = fmt.Fprintf(suggestionsFile, "\t%s\n", r)
		}
		_, _ = fmt.Fprintln(suggestionsFile, ")")
	}
	return nil
}

func createPluginModuleFile(moduleFileName string, module *checks.ModuleInfo) error {
	moduleFile, err := os.Create(moduleFileName)
	if err != nil {
		return err
	}
	//noinspection GoUnhandledErrorResult
	defer moduleFile.Close()

	fmt.Printf("Writing go module file [%s], please use its content to replace your go.mod file\n", moduleFileName)

	// Print out the module
	_, _ = fmt.Fprintf(moduleFile, "module %s\n\n", module.Name)

	// Print out the version
	_, _ = fmt.Fprintf(moduleFile, "go %s\n\n", module.Version)

	// Print out the merged `require` section
	if requires := module.Require; len(requires) > 0 {
		_, _ = fmt.Fprintln(moduleFile, `require (
	// Merged 'require' section of the suggestions and your go.mod file:`)
		keys := getSortedKeys(requires)
		for _, r := range keys {
			_, _ = fmt.Fprintf(moduleFile, "\t%s\n", requires[r])
		}
		_, _ = fmt.Fprintln(moduleFile, ")")
		_, _ = fmt.Fprintln(moduleFile, "")
	}

	// Print out the merged `replace` section
	if replaces := module.Replace; len(replaces) > 0 {
		_, _ = fmt.Fprintln(moduleFile, `replace (
	// Merged 'replace' section of the suggestions and your go.mod file:`)
		keys := getSortedKeys(replaces)
		for _, r := range keys {
			_, _ = fmt.Fprintf(moduleFile, "\t%s\n", replaces[r])
		}
		_, _ = fmt.Fprintln(moduleFile, ")")
	}
	return nil
}

func getSortedKeys(m map[string]string) []string {
	var keys []string
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
