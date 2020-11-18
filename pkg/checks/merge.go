package checks

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/pkg/errors"
)

const (
	None           = ""
	Module         = "module"
	Go             = "go"
	RequireSection = "require"
	ReplaceSection = "replace"
)

type DependencyInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`

	Replacement        bool   `json:"replacement"`
	ReplacementName    string `json:"replacementName,omitempty"`
	ReplacementVersion string `json:"replacementVersion,omitempty"`
}

type Section string

func (s Section) String() string {
	return string(s)
}

type ModuleInfo struct {
	Name    string
	Version string
	Require map[string]string
	Replace map[string]string
}

func MergeModuleFiles(moduleFilePath, glooDepsFilePath string) (*ModuleInfo, error) {
	pluginModule, err := ParseModuleFile(moduleFilePath)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to parse plugin go.mod file")
	}
	gloonDeps, err := ParseDependenciesFile(glooDepsFilePath)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to parse Gloo Enterprise go.mod file")
	}

	merged := mergeModules(pluginModule, gloonDeps)
	return merged, nil
}

func mergeModules(pluginModule *ModuleInfo, glooDependencies map[string]DependencyInfo) *ModuleInfo {
	// create new module with merged require and replace entries
	merged := &ModuleInfo{Name: pluginModule.Name, Version: pluginModule.Version,
		Require: copyMap(pluginModule.Require),
		Replace: copyMap(pluginModule.Replace),
	}

	for name, di := range glooDependencies {
		version := di.Version
		if len(di.ReplacementVersion) > 0 {
			version = di.ReplacementVersion
		}

		replaceName := name
		if len(di.ReplacementName) > 0 {
			replaceName = di.ReplacementName
		}

		merged.Replace[name] = di.Name + " => " + replaceName + " " + version
	}

	//set empty maps to nil
	if len(merged.Replace) == 0 {
		merged.Replace = nil
	}
	if len(merged.Require) == 0 {
		merged.Require = nil
	}

	return merged
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

func ParseModuleFile(filePath string) (*ModuleInfo, error) {
	if err := checkFile(filePath); err != nil {
		return nil, err
	}

	depFile, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	//noinspection GoUnhandledErrorResult
	defer depFile.Close()

	module := &ModuleInfo{}

	scanner := bufio.NewScanner(depFile)
	section := None
	for scanner.Scan() {
		line := scanner.Text()

		depInfo := strings.Fields(line)
		depInfoLen := len(depInfo)

		//skip empty and closing lines
		if depInfoLen <= 1 || strings.HasPrefix(depInfo[0], "//") {
			if depInfoLen == 1 && depInfo[0] == ")" {
				//closing section indicator
				section = None
			}
			continue
		}

		switch section {
		case RequireSection:
			module.Require[depInfo[0]] = strings.TrimSpace(line)
		case ReplaceSection:
			module.Replace[depInfo[0]] = strings.TrimSpace(line)
		default:
			switch depInfo[0] {
			case Module:
				module.Name = depInfo[1]
				continue
			case Go:
				module.Version = depInfo[1]
				continue
			case RequireSection:
				section = RequireSection
				module.Require = make(map[string]string)
				continue
			case ReplaceSection:
				section = ReplaceSection
				module.Replace = make(map[string]string)
				continue
			default:
				if depInfo[1] == "(" {
					return nil, fmt.Errorf("unkown section: [%s]. "+
						"Expected on of 'module | go | require | replace'", line)
				}
			}
		}

	}
	return module, scanner.Err()
}

func copyMap(m map[string]string) map[string]string {
	cp := make(map[string]string)
	for k, v := range m {
		cp[k] = v
	}

	return cp
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
