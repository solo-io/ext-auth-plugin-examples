package main

import (
	"encoding/json"
	"fmt"
	"github.com/solo-io/ext-auth-plugin-examples/pkg/checks"
	"io/ioutil"
	"os"
)

const (
	errorReportFile = "mismatched_dependencies.json"
	suggestionsFile = "suggestions"
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

	reportBytes, err := json.MarshalIndent(nonMatchingDeps, "", "  ")
	if err != nil {
		fmt.Printf("Failed to marshall error report to JSON: %s/n", err.Error())
		os.Exit(1)
	}

	fmt.Println(string(reportBytes))

	fmt.Printf("Writing error report file [%s]\n", errorReportFile)
	if err := ioutil.WriteFile(errorReportFile, reportBytes, 0644); err != nil {
		fmt.Printf("Failed to write error report file: %s/n", err.Error())
	}

	overrideFile, err := os.Create(suggestionsFile)
	if err != nil {
		fmt.Printf("Failed to create override file: %s/n", err.Error())
		os.Exit(1)
	}
	//noinspection GoUnhandledErrorResult
	defer overrideFile.Close()

	for _, pair := range nonMatchingDeps {
		switch pair.MismatchType {
		case checks.Require:
			_, _ = fmt.Fprintf(overrideFile, "%s %s\n", pair.Gloo.Name, pair.Gloo.Version)
		case checks.PluginMissingReplace:
			_, _ = fmt.Fprintf(overrideFile, "%s %s => %s %s\n", pair.Gloo.Name, pair.Gloo.Version, pair.Gloo.ReplacementName, pair.Gloo.ReplacementVersion)
		case checks.ReplaceMismatch:
			_, _ = fmt.Fprintf(overrideFile, "%s %s => %s %s\n", pair.Gloo.Name, pair.Gloo.Version, pair.Gloo.ReplacementName, pair.Gloo.ReplacementVersion)
		}
		// Nothing to suggest for the PluginExtraReplace errors
	}

	fmt.Printf("Writing suggestions file [%s], please use its content to update your go.mod file\n", suggestionsFile)

	os.Exit(1)
}
