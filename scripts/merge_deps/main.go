package main

import (
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"github.com/solo-io/ext-auth-plugin-examples/pkg/checks"
	"io/ioutil"
	"os"
	"sort"
	"strconv"
)

const (
	errorReportFile     = "mismatched_dependencies.json"
	suggestionsFileName = "suggestions"
	moduleFileName      = "go.mod"
	backupDirName       = "tmp"
)

func main() {
	if len(os.Args) != 4 {
		fmt.Printf("Must provide 3 arguments: \n\t- plugin go.mod file path \n\t- Glooe go.mod file path\n\t- merge attempts plugin go.mod file\n")
		os.Exit(1)
	}

	pluginsDependenciesFilePath := os.Args[1]
	glooDependenciesFilePath := os.Args[2]
	var (
		mergeAttempt    int
		nonMatchingDeps []checks.DependencyInfoPair
		moduleInfo      *checks.ModuleInfo
		err             error
	)
	if mergeAttempt, err = strconv.Atoi(os.Args[3]); err != nil {
		fmt.Printf("Provided 3th arguments is not a number\n")
		os.Exit(1)
	}

	if moduleInfo, err = checks.ParseModuleInfo(moduleFileName); err != nil {
		fmt.Printf("Failed to read plugin module info: %s/n", err.Error())
		os.Exit(1)
	}

	var mergedDeps map[string]checks.DependencyInfo

	for i := 1; mergeAttempt > 0 && i <= mergeAttempt; i++ {
		if nonMatchingDeps, err = checks.CompareDependencies(pluginsDependenciesFilePath, glooDependenciesFilePath); err != nil {
			fmt.Printf("Failed to compare dependencies: %s/n", err.Error())
			os.Exit(1)
		}

		if len(nonMatchingDeps) == 0 {
			fmt.Println("All shared dependencies match")
			os.Exit(0)
		}
		fmt.Println("Plugin and Gloo Enterprise dependencies do not match!")
		if i < mergeAttempt {
			fmt.Printf("Adjusting Plugin dependencies and start comparing again (%d times)\n", i)
		}

		if mergedDeps, err = mergeDependencies(pluginsDependenciesFilePath, nonMatchingDeps); err != nil {
			fmt.Printf("Failed to merge non matching dependencies: %s/n", err.Error())
			os.Exit(1)
		}

		if err = backupPluginModuleFile(i); err != nil {
			fmt.Printf("Failed to backup current '%s' file: %s/n", moduleFileName, err.Error())
			os.Exit(1)
		}

		if err = createPluginModuleFile(moduleInfo, mergedDeps); err != nil {
			fmt.Printf("Failed to write new merged '%s' file: %s/n", moduleFileName, err.Error())
			os.Exit(1)
		}
	}

	// 1. Write the report to stdout
	reportBytes, err := json.MarshalIndent(nonMatchingDeps, "", "  ")
	if err != nil {
		fmt.Printf("Failed to marshall error report to JSON: %s/n", err.Error())
		os.Exit(1)
	}
	fmt.Println(string(reportBytes))

	// 2. Write the report to a file
	fmt.Printf("Writing error report file [%s]\n", errorReportFile)
	if err := ioutil.WriteFile(errorReportFile, reportBytes, 0644); err != nil {
		fmt.Printf("Failed to write error report file: %s/n", err.Error())
	}

	// 3. Create a file with suggested changes to go.mod
	if err := createSuggestionsFile(nonMatchingDeps); err != nil {
		fmt.Printf("Failed to create suggestions file: %s/n", err.Error())
	}
	os.Exit(1)
}

func backupPluginModuleFile(suffix int) error {
	if err := os.MkdirAll(backupDirName, os.ModePerm); err != nil {
		return err
	}
	backupModuleFileName := fmt.Sprintf("%s/%s-%d", backupDirName, moduleFileName, suffix)
	return os.Rename(moduleFileName, backupModuleFileName)
}

func mergeDependencies(pluginsDependenciesFilePath string, nonMatchingDeps []checks.DependencyInfoPair) (map[string]checks.DependencyInfo, error) {
	pluginDependencies, err := checks.ParseDependenciesFile(pluginsDependenciesFilePath)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to parse plugin go.mod file")
	}
	for _, dep := range nonMatchingDeps {
		depInfo := pluginDependencies[dep.Plugin.Name]
		depInfo.Version = dep.Gloo.Version
		if dep.Gloo.Replacement {
			depInfo.Version = dep.Gloo.ReplacementVersion
		}
	}
	return pluginDependencies, nil
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

func createPluginModuleFile(info *checks.ModuleInfo, dependencies map[string]checks.DependencyInfo) error {
	moduleFile, err := os.Create(moduleFileName)
	if err != nil {
		return err
	}
	//noinspection GoUnhandledErrorResult
	defer moduleFile.Close()

	fmt.Printf("Writing go module file [%s], please use its content to replace your go.mod file\n", moduleFileName)

	// Print out the module
	_, _ = fmt.Fprintf(moduleFile, "module %s\n", info.Name)

	// Print out the version
	_, _ = fmt.Fprintf(moduleFile, "go %s\n", info.Version)

	var dep checks.DependencyInfo
	// Print out the merged `require` section
	_, _ = fmt.Fprintln(moduleFile, `require (
	// Merged 'require' section of the suggestions and your go.mod file:`)
	keys := getSortedKeys(dependencies)
	for _, r := range keys {
		if dep = dependencies[r]; !dep.Replacement {
			_, _ = fmt.Fprintf(moduleFile, "\t%s\n", fmt.Sprintf("%s %s", dep.Name, dep.Version))
		}
	}
	_, _ = fmt.Fprintln(moduleFile, ")")

	// Print out the merged `replace` section
	_, _ = fmt.Fprintln(moduleFile, `replace (
	// Merged 'replace' section of the suggestions and your go.mod file:`)
	keys = getSortedKeys(dependencies)
	for _, r := range keys {
		if dep = dependencies[r]; dep.Replacement {
			_, _ = fmt.Fprintf(moduleFile, "\t%s\n",
				fmt.Sprintf("%s %s => %s %s", dep.Name, dep.Version, dep.ReplacementName, dep.ReplacementVersion))
		}
	}
	_, _ = fmt.Fprintln(moduleFile, ")")
	return nil
}

func getSortedKeys(m map[string]checks.DependencyInfo) []string {
	var keys []string
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
