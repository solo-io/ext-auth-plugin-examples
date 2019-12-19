package main

import (
	"encoding/json"
	"fmt"
	"github.com/solo-io/ext-auth-plugin-examples/pkg/checks"
	"io/ioutil"
	"os"
)

const errorReportFile = "mismatched_dependencies.json"

func main() {
	if len(os.Args) != 3 {
		fmt.Printf("Must provide 2 arguments: \n\t- plugin go.mod file path \n\t- Glooe go.mod file path\n")
		os.Exit(1)
	}

	pluginGoModFilePath := os.Args[1]
	glooGoModFilePath := os.Args[2]

	report, err := checks.CompareDependencies(pluginGoModFilePath, glooGoModFilePath)
	if err != nil {
		fmt.Printf("Failed to compare dependencies: %s/n", err.Error())
		os.Exit(1)
	}

	if len(report) == 0 {
		fmt.Println("All shared dependencies match")
		os.Exit(0)
	}

	fmt.Println("Plugin and Gloo Enterprise dependencies do not match!")

	reportBytes, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		fmt.Printf("Failed to marshall error report to JSON: %s/n", err.Error())
		os.Exit(1)
	}

	fmt.Println(string(reportBytes))

	fmt.Printf("Writing error report file [%s]\n", errorReportFile)
	if err := ioutil.WriteFile(errorReportFile, reportBytes, 0644); err != nil {
		fmt.Printf("Failed to write error report file: %s/n", err.Error())
	}
	os.Exit(1)
}
