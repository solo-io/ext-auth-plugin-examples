package checks

import (
	"github.com/pkg/errors"
	"golang.org/x/mod/modfile"
	"golang.org/x/mod/module"
	"io/ioutil"
	"os"
)

// Used to easily look up module info by module name.
type indexedModFile struct {
	required,
	requiredWithoutReplace map[string]*module.Version
	replaced map[string]*modfile.Replace
}

func CompareDependencies(pluginGoModFilePath, glooGoModFilePath string) (Report, error) {

	pluginGoModFile, err := parseModFile(pluginGoModFilePath)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to parse plugin go.mod file")
	}
	glooGoModFile, err := parseModFile(glooGoModFilePath)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to parse  Gloo Enterprise go.mod file")
	}

	rep := Report{}

	// Start by checking all the [require] entries in the plugin go.mod file that don't have a corresponding
	// [replace] entry in the same file. We will check [replace] entries in the plugin go.mod file later.
	for moduleName, modInfo := range pluginGoModFile.requiredWithoutReplace {

		// Check if GlooE defines a replacement for the module
		if correspondingGlooReplace, ok := glooGoModFile.replaced[modInfo.Path]; ok {

			// If it does, check the name of the replacement
			if correspondingGlooReplace.New.Path != moduleName {
				rep.AddEntry(GlooReplaceNameMismatch, moduleName,
					glooDefinesReplacementWithDifferentName(moduleName, correspondingGlooReplace.New.Path))
				continue
			}

			// If the name matches, check the version
			if correspondingGlooReplace.New.Version != modInfo.Version {
				rep.AddEntry(GlooReplaceVersionMismatch, moduleName,
					glooDefinesReplacementWithDifferentVersion(moduleName, modInfo.Version, correspondingGlooReplace.New.Version))
				continue
			}
		}

		// Check if GlooE defines a requirement for the module
		if correspondingGlooRequire, ok := glooGoModFile.required[modInfo.Path]; ok {
			// If it does, check that the versions align
			if correspondingGlooRequire.Version != modInfo.Version {
				rep.AddEntry(GlooRequireVersionMismatch, moduleName,
					glooDefinesRequirementWithDifferentVersion(moduleName, modInfo.Version, correspondingGlooRequire.Version))
			}
		}
	}

	// Now check all the [replace] entries in the plugin go.mod file.
	// We do not want to [replace] any [require] (without a [replace]) that appears in the GlooE go.mod file.
	for moduleName, modInfo := range pluginGoModFile.replaced {

		// Check if GlooE defines a requirement (and no replacement) for the module.
		if correspondingGlooRequire, ok := glooGoModFile.requiredWithoutReplace[modInfo.Old.Path]; ok {
			rep.AddEntry(PluginReplaceIsGlooRequirement, moduleName,
				pluginDefinesReplacementThatIsRequirementInGloo(moduleName, correspondingGlooRequire.Version))
			continue
		}
	}

	// Lastly, verify that all the [replace] entries in the GlooE go.mod file appear in the plugin go.mod file as well.
	// We need this in order to guarantee that the indirect dependencies that are shared will match.
	for moduleName, modInfo := range glooGoModFile.replaced {

		correspondingPluginReplace, ok := pluginGoModFile.replaced[modInfo.Old.Path]
		if !ok {
			rep.AddEntry(PluginMissingReplace, moduleName,
				pluginIsMissingReplacement(moduleName, modInfo.New.String()))
			continue
		}
		if correspondingPluginReplace.New.Path != modInfo.New.Path || correspondingPluginReplace.New.Version != modInfo.New.Version {
			rep.AddEntry(PluginReplaceMismatch, moduleName,
				glooDefinesReplacementThatDoesNotMatchInPlugin(moduleName, correspondingPluginReplace.New.String(), modInfo.New.String()))
		}
	}

	return rep, nil
}

func parseModFile(filePath string) (*indexedModFile, error) {
	if err := checkFile(filePath); err != nil {
		return nil, err
	}

	fileContentsBytes, err := ioutil.ReadFile(filePath)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to load go.mod file from path [%s]", filePath)
	}

	modFile, err := modfile.Parse(filePath, fileContentsBytes, nil)
	if err != nil {
		return nil, err
	}

	indexedFile := &indexedModFile{
		required:               make(map[string]*module.Version),
		requiredWithoutReplace: make(map[string]*module.Version),
		replaced:               make(map[string]*modfile.Replace),
	}
	for _, requirement := range modFile.Require {
		indexedFile.required[requirement.Mod.Path] = &requirement.Mod
	}
	for _, replacement := range modFile.Replace {
		indexedFile.replaced[replacement.Old.Path] = replacement
	}
	for _, requirement := range modFile.Require {
		if _, ok := indexedFile.replaced[requirement.Mod.Path]; !ok {
			indexedFile.requiredWithoutReplace[requirement.Mod.Path] = &requirement.Mod
		}
	}

	return indexedFile, nil
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
