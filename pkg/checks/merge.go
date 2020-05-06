package checks

import (
	"bufio"
	"fmt"
	"github.com/pkg/errors"
	"os"
	"strings"
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
	pluginModule, err := parseModuleFile(moduleFilePath)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "failed to parse plugin go.mod file")
	}
	glooModule, err := parseDependenciesFile(glooDepsFilePath)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "failed to parse  Gloo Enterprise go.mod file")
	}

	merged := &ModuleInfo{Name: pluginModule.Name, Version: pluginModule.Version,
		Require: mergeMaps(pluginModule.Require, glooModule.Require),
		Replace: mergeMaps(pluginModule.Replace, glooModule.Replace),
	}
	pluginDeps, err := toDependencyInfo(merged)
	gloonDeps, err := toDependencyInfo(glooModule)

	nonMatchingDeps := compareDependencies(pluginDeps, gloonDeps)

	return merged, nonMatchingDeps, err
}

func parseDependenciesFile(filePath string) (*ModuleInfo, error) {
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

func parseModuleFile(filePath string) (*ModuleInfo, error) {
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
	if m == nil {
		return nil
	}
	cp := make(map[string]string)
	for k, v := range m {
		cp[k] = v
	}

	return cp
}

func mergeMaps(base, overrides map[string]string) map[string]string {
	if base == nil && overrides == nil {
		return nil
	}
	var m map[string]string
	if base != nil {
		m = copyMap(base)
		if overrides != nil {
			for k, v := range overrides {
				m[k] = v
			}
		}
	} else {
		m = copyMap(overrides)
	}

	return m
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
