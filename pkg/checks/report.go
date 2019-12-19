package checks

import "fmt"

type DependencyErrorType int

const (
	// Occurs when the PLUGIN has a [require], but GLOO has a [replace] for the same module with a different name.
	GlooReplaceNameMismatch DependencyErrorType = iota
	// Occurs when the PLUGIN has a [require], but GLOO has a [replace] for the same module with a different version.
	GlooReplaceVersionMismatch
	// Occurs when the PLUGIN has a [require], but GLOO has a [require] for the same module with a different version.
	GlooRequireVersionMismatch
	// Occurs when the PLUGIN has a [replace], but GLOO has a [require] (but no [replace]) for the same module.
	PluginReplaceIsGlooRequirement
	// Occurs when GLOO has a [replace] for a module, but the PLUGIN does not.
	PluginMissingReplace
	// Occurs when GLOO has a [replace] for a module, but the PLUGIN has a [replace] for the same module that does not match.
	PluginReplaceMismatch
)

var (
	glooDefinesReplacementWithDifferentName = func(pluginModName, glooModName string) string {
		return fmt.Sprintf("your plugin requires the [%s] module, but GlooE replaces "+
			"that module with one named [%s]. Please add a 'replace' entry for this module in your go.mod file "+
			"to matche the GlooE one", pluginModName, glooModName)
	}
	glooDefinesReplacementWithDifferentVersion = func(pluginModName, pluginModVersion, glooModVersion string) string {
		return fmt.Sprintf("your plugin requires the [%s] module with version [%s], "+
			"but GlooE replaces that module with one that has the same name and version [%s]. Please add a "+
			"'replace' entry for this module in your go.mod file to match the GlooE one", pluginModName,
			pluginModVersion, glooModVersion)
	}
	glooDefinesRequirementWithDifferentVersion = func(pluginModName, pluginModVersion, glooModVersion string) string {
		return fmt.Sprintf("your plugin requires the [%s] module with version [%s], "+
			"but GlooE requires that module with version [%s]. Please update the version of this module in your go.mod "+
			"file to matches the GlooE one", pluginModName,
			pluginModVersion, glooModVersion)
	}
	pluginDefinesReplacementThatIsRequirementInGloo = func(pluginModName, glooRequirementVersion string) string {
		return fmt.Sprintf("your plugin defines a replacement for the [%s] module, but GlooE has a [require] "+
			"entry for the same module with version [%s] (and no [replace] for it). Please remove the replacement in "+
			"your go.mod and instead use a [require] entry that matches the GlooE one", pluginModName, glooRequirementVersion)
	}
	pluginIsMissingReplacement = func(pluginModName, glooReplacement string) string {
		return fmt.Sprintf("GlooE defines a replacement for the [%s] module but your plugin does not. "+
			"Please add a [replace] entry to your go.mod file to replace the module with [%s]", pluginModName, glooReplacement)
	}
	glooDefinesReplacementThatDoesNotMatchInPlugin = func(pluginModName, pluginReplacement, glooReplacement string) string {
		return fmt.Sprintf("GlooE defines a replacement for the [%s] module that does not match the "+
			"replacement for the same module in your plugin. The two [replace] entries must be identical. "+
			"Your plugin has: %s, GlooE has %s", pluginModName, pluginReplacement, glooReplacement)
	}
)

type DependencyError struct {
	Kind    DependencyErrorType `json:"-"`
	Module  string              `json:"module"`
	Message string              `json:"message"`
}

type Report map[string]*DependencyError

func (r Report) AddEntry(kind DependencyErrorType, moduleName, message string) {
	// Only add one error per module, multiple errors should point to the same cause
	if _, ok := r[moduleName]; !ok {
		r[moduleName] = &DependencyError{
			Kind:    kind,
			Module:  moduleName,
			Message: message,
		}
	}
}

func (r Report) GetEntry(moduleName string) *DependencyError {
	if depErrors, ok := r[moduleName]; ok {
		return depErrors
	}
	return nil
}
