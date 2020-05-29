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

func MergeModuleFiles(moduleFilePath, glooDepsFilePath string) (*ModuleInfo, []DependencyInfoPair, error) {
	pluginModule, err := ParseModuleFile(moduleFilePath)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "failed to parse plugin go.mod file")
	}
	glooModule, err := ParseDependenciesFile(glooDepsFilePath)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "failed to parse Gloo Enterprise go.mod file")
	}

	merged := mergeModules(pluginModule, glooModule)
	pluginDeps, err := toDependencyInfo(merged)
	gloonDeps, err := toDependencyInfo(glooModule)

	nonMatchingDeps := CompareDependencies(pluginDeps, gloonDeps)

	return merged, nonMatchingDeps, err
}

func mergeModules(pluginModule, glooModule *ModuleInfo) *ModuleInfo {
	// create new module with merged require and replace entries
	merged := &ModuleInfo{Name: pluginModule.Name, Version: pluginModule.Version,
		Require: copyMap(pluginModule.Require),
		Replace: copyMap(pluginModule.Replace),
	}

	for name := range pluginModule.Require {
		// pin dependency to the same version as the Gloo one using a [require] clause
		if glooEquivalent, ok := glooModule.Require[name]; ok {
			merged.Require[name] = glooEquivalent
			continue
		}
		// add [replace] clause matching the Gloo one
		if glooEquivalent, ok := glooModule.Replace[name]; ok {
			merged.Replace[name] = glooEquivalent
			continue
		}
	}

	for name, replace := range pluginModule.Replace {
		// remove the [replace] clause and pin your dependency to the same version as the Gloo one using a [require] clause
		if glooEquivalent, ok := glooModule.Require[name]; ok {
			merged.Require[name] = glooEquivalent
			// gloo require entries are not allowed to be replaced
			delete(merged.Replace, name)
			continue
		}
		// update [replace] clause matching the Gloo one if version is specified
		if glooEquivalent, ok := glooModule.Replace[name]; ok && len(strings.Fields(replace)) == 5 {
			merged.Replace[name] = glooEquivalent
			continue
		}
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

func ParseDependenciesFile(filePath string) (*ModuleInfo, error) {
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
	skippedFirstLine := false
	for scanner.Scan() {
		line := scanner.Text()

		depInfo := strings.Fields(line)

		// First line is the name of the module the `go list -m all` command was ran for
		if !skippedFirstLine && len(depInfo) == 1 {
			module.Name = depInfo[0]
			skippedFirstLine = true
			continue
		}

		switch len(depInfo) {
		case 2:
			if module.Require == nil {
				module.Require = make(map[string]string)
			}
			module.Require[depInfo[0]] = strings.TrimSpace(line)
		case 5:
			if module.Replace == nil {
				module.Replace = make(map[string]string)
			}
			module.Replace[depInfo[0]] = strings.TrimSpace(line)
		default:
			return nil, errors.Errorf("malformed dependency: [%s]. "+
				"Expected format 'NAME VERSION' or 'NAME VERSION => REPLACE_NAME REPLACE_VERSION'", line)
		}
	}
	return module, scanner.Err()
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

func toDependencyInfo(module *ModuleInfo) (map[string]DependencyInfo, error) {
	dis := map[string]DependencyInfo{}
	//always add replacements
	for name, replace := range module.Replace {
		depInfo := strings.Fields(replace)
		switch len(depInfo) {
		case 4:
			dis[name] = DependencyInfo{
				Name:               depInfo[0],
				Version:            depInfo[1],
				Replacement:        true,
				ReplacementName:    depInfo[2],
				ReplacementVersion: depInfo[3],
			}
		case 5:
			dis[name] = DependencyInfo{
				Name:               depInfo[0],
				Version:            depInfo[1],
				Replacement:        true,
				ReplacementName:    depInfo[3],
				ReplacementVersion: depInfo[4],
			}
		default:
			return nil, errors.Errorf("malformed replace dependency: [%s]. "+
				"Expected format 'NAME VERSION => REPLACE_NAME REPLACE_VERSION'", replace)
		}
	}

	//only add requires if key does not exists
	for name, require := range module.Require {
		if _, present := dis[name]; present {
			continue
		}

		depInfo := strings.Fields(require)
		switch len(depInfo) {
		case 2, 3:
			dis[name] = DependencyInfo{
				Name:    depInfo[0],
				Version: depInfo[1],
			}
		default:
			return nil, errors.Errorf("malformed require dependency: [%s]. "+
				"Expected format 'NAME VERSION'", require)
		}
	}

	return dis, nil
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
