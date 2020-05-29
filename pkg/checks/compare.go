package checks

import (
	"os"
	"strconv"
)

type MismatchType int

const (
	isForked = "IS_FORKED"

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

func CompareDependencies(pluginDependencies, glooDependencies map[string]DependencyInfo) []DependencyInfoPair {
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
		// by using this hack, we are able to support forked repos
		if isSet, _ := strconv.ParseBool(os.Getenv(isForked)); isSet {
			return true, Ok, ""
		}
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
