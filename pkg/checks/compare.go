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

	Module          = "module"
	Go              = "go"
	RequireSection  = "require"
	ReplaceSection  = "replace"
)

type Section string

func (s Section) String() string {
	return string(s)
}

type ModuleInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

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

	pluginDependencies, err := ParseDependenciesFile(pluginsDepsFilePath)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to parse plugin go.mod file")
	}
	glooDependencies, err := ParseDependenciesFile(glooDepsFilePath)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to parse  Gloo Enterprise go.mod file")
	}

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

	return nonMatchingDeps, nil
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

func ParseDependenciesFile(filePath string) (map[string]DependencyInfo, error) {
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

func ParseModuleInfo(filePath string) (*ModuleInfo, error) {
	if err := checkFile(filePath); err != nil {
		return nil, err
	}

	depFile, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	//noinspection GoUnhandledErrorResult
	defer depFile.Close()

	moduleInfo := &ModuleInfo{}

	scanner := bufio.NewScanner(depFile)
	for scanner.Scan() {
		line := scanner.Text()

		depInfo := strings.Fields(line)
		depInfoLen := len(depInfo)

		//skip empty and closing lines
		if depInfoLen <= 1 || depInfo[0] == "//" {
			continue
		}

		switch depInfo[0] {
		case Module:
			moduleInfo.Name = depInfo[1]
			continue
		case Go:
			moduleInfo.Version = depInfo[1]
		case RequireSection, ReplaceSection:
			// stop the loop , we are passed the info sections
			break
		}
	}
	return moduleInfo, scanner.Err()

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
