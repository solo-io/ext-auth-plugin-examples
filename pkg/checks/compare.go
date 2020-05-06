package checks

import (
	"bufio"
	"github.com/pkg/errors"
	"os"
	"strings"
)

type MismatchType int

const (
	Ok MismatchType = iota
	Require
	PluginMissingReplace
	PluginExtraReplace
	ReplaceMismatch
	Ko
)

type DependencyInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`

	Replacement        bool   `json:"replacement"`
	ReplacementName    string `json:"replacementName,omitempty"`
	ReplacementVersion string `json:"replacementVersion,omitempty"`
}

type DependencyInfoPair struct {
	Message      string         `json:"message"`
	MismatchType MismatchType   `json:"-"`
	Plugin       DependencyInfo `json:"pluginDependencies"`
	Gloo         DependencyInfo `json:"glooDependencies"`
}

func CompareDependencies(pluginsDepsFilePath, glooDepsFilePath string) ([]DependencyInfoPair, error) {

	pluginDependencies, err := parseDependenciesFile_Deprecated(pluginsDepsFilePath)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to parse plugin go.mod file")
	}
	glooDependencies, err := parseDependenciesFile_Deprecated(glooDepsFilePath)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to parse  Gloo Enterprise go.mod file")
	}

	nonMatchingDeps := compareDependencies(pluginDependencies, glooDependencies)

	return nonMatchingDeps, nil
}

func compareDependencies(pluginDependencies map[string]DependencyInfo, glooDependencies map[string]DependencyInfo) []DependencyInfoPair {
	var nonMatchingDeps []DependencyInfoPair
	for name, depInfo := range pluginDependencies {

		// Just check libraries that are shared with GlooE
		if glooEquivalent, ok := glooDependencies[name]; ok {
			if match, mismatchType, msg := matches(glooEquivalent, depInfo); !match {
				nonMatchingDeps = append(nonMatchingDeps, DependencyInfoPair{
					Message:      msg,
					MismatchType: mismatchType,
					Plugin:       depInfo,
					Gloo:         glooEquivalent,
				})
			}
		}
	}
	return nonMatchingDeps
}

func matches(glooDep, pluginDep DependencyInfo) (bool, MismatchType, string) {
	// If both are simple dependencies, just compare the versions
	switch {
	case glooDep.Replacement == false && pluginDep.Replacement == false:
		if glooDep.Version == pluginDep.Version {
			return true, Ok, ""
		} else {
			return false, Require, "Please pin your dependency to the same version as the Gloo one using a [require] clause"
		}
	case glooDep.Replacement == true && pluginDep.Replacement == false:
		return false, PluginMissingReplace, "Please add a [replace] clause matching the Gloo one"
	case glooDep.Replacement == false && pluginDep.Replacement == true:
		return false, PluginExtraReplace, "Please remove the [replace] clause and pin your dependency to the same version as the Gloo one using a [require] clause"
	case glooDep.Replacement && pluginDep.Replacement:
		if glooDep.ReplacementName == pluginDep.ReplacementName && glooDep.ReplacementVersion == pluginDep.ReplacementVersion {
			return true, Ok, ""
		} else {
			return false, ReplaceMismatch, "The plugin [replace] clause must match the Gloo one"
		}
	}

	return false, Ko, "internal error"
}

func parseDependenciesFile_Deprecated(filePath string) (map[string]DependencyInfo, error) {
	if err := checkFile(filePath); err != nil {
		return nil, err
	}

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
