package main

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"

	"github.com/solo-io/ext-auth-plugin-examples/pkg/checks"
)

func main() {
	if len(os.Args) != 3 {
		fmt.Printf("Must provide 2 arguments: \n\t- Plugin go.mod file path\n\t- Glooe dependencies file path\n")
		os.Exit(1)
	}

	pluginsModuleFilePath := os.Args[1]
	glooDependenciesFilePath := os.Args[2]
	var (
		nonMatchingDeps []checks.DependencyInfoPair
		mergedModule    *checks.ModuleInfo
		err             error
	)
	if mergedModule, nonMatchingDeps, err = checks.MergeModuleFiles(pluginsModuleFilePath, glooDependenciesFilePath); err != nil {
		fmt.Printf("Failed to resolve dependencies: %s\n", err.Error())
		os.Exit(1)
	}

	if len(nonMatchingDeps) == 0 {
		fmt.Printf("All shared dependencies match, writing new merged '%s'\n", pluginsModuleFilePath)

		if err = createPluginModuleFile(pluginsModuleFilePath, mergedModule); err != nil {
			fmt.Printf("failed to write new merged '%s' file: %s\n", pluginsModuleFilePath, err.Error())
			os.Exit(1)
		}
		os.Exit(0)
	}
	fmt.Println("Plugin and Gloo Enterprise dependencies do not match after merge")

	// 1. Write the report to stdout
	reportBytes, err := json.MarshalIndent(nonMatchingDeps, "", "  ")
	if err != nil {
		fmt.Printf("Failed to marshall error report to JSON: %s/n", err.Error())
		os.Exit(1)
	}
	fmt.Println(string(reportBytes))

	os.Exit(1)
}

func createPluginModuleFile(moduleFileName string, module *checks.ModuleInfo) error {
	moduleFile, err := os.Create(moduleFileName)
	if err != nil {
		return err
	}
	//noinspection GoUnhandledErrorResult
	defer moduleFile.Close()

	fmt.Printf("Writing go module file [%s], its content will replace your go.mod file\n", moduleFileName)

	// Print out the module
	_, _ = fmt.Fprintf(moduleFile, "module %s\n\n", module.Name)

	// Print out the version
	_, _ = fmt.Fprintf(moduleFile, "go %s\n\n", module.Version)

	// Print out the merged `require` section
	if requires := module.Require; len(requires) > 0 {
		_, _ = fmt.Fprintln(moduleFile, `require (
	// Merged 'require' section of the Gloo depenencies and your go.mod file:`)
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
	// Merged 'replace' section of the Gloo depenencies and your go.mod file:`)
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
