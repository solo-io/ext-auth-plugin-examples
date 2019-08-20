package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/pelletier/go-toml"
)

const errorReportFile = "mismatched_dependencies.json"

type depProjects map[string]dependencyInfo

// `Revision` is always present on a project stanza. Additionally, either `branch` or `version` can be present.
// Check the [dep docs](https://golang.github.io/dep/docs/Gopkg.lock.html#projects) for full reference.
type dependencyInfo struct {
	Name     string `json:"name"`
	Version  string `json:"version,omitempty"`
	Revision string `json:"revision"`
	Branch   string `json:"branch,omitempty"`
}

type dependencyInfoPair struct {
	Plugin dependencyInfo `json:"pluginDependencies"`
	GlooE  dependencyInfo `json:"glooeDependencies"`
}

func (d dependencyInfo) matches(that dependencyInfo) bool {
	// `Revision` is the ultimate source of truth, `version` or `branch` are potentially floating references
	if d.Revision == that.Revision {
		return true
	}
	return false
}

func main() {

	if len(os.Args) != 3 {
		fmt.Printf("Must provide 2 arguments: \n\t- plugin Gopkg.lock file path \n\t- Glooe Gopkg.lock file path\n")
		os.Exit(1)
	}

	pluginDepLockFile := os.Args[1]
	glooeDepLockFile := os.Args[2]

	if err := checkFile(pluginDepLockFile); err != nil {
		fmt.Printf("%s\n", err.Error())
		os.Exit(1)
	}

	if err := checkFile(glooeDepLockFile); err != nil {
		fmt.Printf("%s\n", err.Error())
		os.Exit(1)
	}

	pluginDependencies, err := getDependencies(pluginDepLockFile)
	if err != nil {
		fmt.Printf("Failed to get plugin dependencies: %s/n", err.Error())
		os.Exit(1)
	}
	glooeDependencies, err := getDependencies(glooeDepLockFile)
	if err != nil {
		fmt.Printf("Failed to get GlooE dependencies: %s/n", err.Error())
		os.Exit(1)
	}

	var nonMatchingDeps []dependencyInfoPair
	for name, depInfo := range pluginDependencies {

		// Just check libraries that are shared with GlooE
		if glooeEquivalent, ok := glooeDependencies[name]; ok {
			if !glooeEquivalent.matches(depInfo) {
				nonMatchingDeps = append(nonMatchingDeps, dependencyInfoPair{
					Plugin: depInfo,
					GlooE:  glooeEquivalent,
				})
			}
		}
	}

	if len(nonMatchingDeps) == 0 {
		fmt.Println("All shared dependencies match")
		os.Exit(0)
	}

	fmt.Printf("Plugin and GlooE dependencies do not match!\n")

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
	os.Exit(1)
}

func getDependencies(depLockFilePath string) (depProjects, error) {
	pluginDependencies, err := parseGoPkgLock(depLockFilePath)
	if err != nil {
		return nil, err
	}
	return collectDependencyInfo(pluginDependencies), nil
}

func parseGoPkgLock(path string) ([]*toml.Tree, error) {
	config, err := toml.LoadFile(path)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("failed to parse %s file: %s", path, err.Error()))
	}

	tomlTree := config.Get("projects")

	switch typedTree := tomlTree.(type) {
	case []*toml.Tree:
		return typedTree, nil
	default:
		return nil, fmt.Errorf("unable to parse toml tree")
	}
}

func collectDependencyInfo(deps []*toml.Tree) depProjects {
	dependencies := make(depProjects)

	for _, t := range deps {

		name := t.Get("name").(string)

		info := dependencyInfo{
			Name:     name,
			Revision: t.Get("revision").(string),
		}

		if version, ok := t.Get("version").(string); ok {
			info.Version = version
		}
		if branch, ok := t.Get("branch").(string); ok {
			info.Branch = branch
		}

		dependencies[name] = info
	}
	return dependencies
}

func checkFile(filename string) error {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return errors.New(filename + " file not found")
	}
	if info.IsDir() {
		return errors.New(filename + " is a directory")
	}
	return nil
}
