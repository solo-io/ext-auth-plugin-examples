package main

import (
	"encoding/json"
	"fmt"
	"github.com/solo-io/ext-auth-plugin-examples/pkg/checks"
	"io/ioutil"
	"os"
)

const (
	errorReportFile     = "mismatched_dependencies.json"
	suggestionsFileName = "suggestions"
)

func main() {
	if len(os.Args) != 3 {
		fmt.Printf("Must provide 2 arguments: \n\t- plugin go.mod file path \n\t- Glooe go.mod file path\n")
		os.Exit(1)
	}

	pluginsDependenciesFilePath := os.Args[1]
	glooDependenciesFilePath := os.Args[2]

	nonMatchingDeps, err := checks.CompareDependencies(pluginsDependenciesFilePath, glooDependenciesFilePath)
	if err != nil {
		fmt.Printf("Failed to compare dependencies: %s/n", err.Error())
		os.Exit(1)
	}

	if len(nonMatchingDeps) == 0 {
		fmt.Println("All shared dependencies match")
		os.Exit(0)
	}

	fmt.Println("Plugin and Gloo Enterprise dependencies do not match!")

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
			suggestionMap[checks.ReplaceMismatch] = append(suggestionMap[checks.Require],
				fmt.Sprintf("%s %s => %s %s", pair.Gloo.Name, pair.Gloo.Version, pair.Gloo.ReplacementName, pair.Gloo.ReplacementVersion))
		}
	}

	// Print out the suggested changes for the `require` section of the go.mod file
	if requires, ok := suggestionMap[checks.Require]; ok && len(requires) > 0 {
		_, _ = fmt.Fprintln(suggestionsFile, `require (
	// You other requirements (remove the ones that collide with the following suggestions)...`)
		for _, r := range requires {
			_, _ = fmt.Fprintf(suggestionsFile, "\t%s", r)
		}
		_, _ = fmt.Fprintln(suggestionsFile, ")")
	}

	// Print out the suggested changes for the `replace` section of the go.mod file

	if replacements, ok := suggestionMap[checks.ReplaceMismatch]; ok && len(replacements) > 0 {
		_, _ = fmt.Fprintln(suggestionsFile, `replace (
	// You other replacements (remove the ones that collide with the following suggestions)...`)
		for _, r := range replacements {
			_, _ = fmt.Fprintf(suggestionsFile, "\t%s", r)
		}
		_, _ = fmt.Fprintln(suggestionsFile, ")")
	}
	return nil
}
