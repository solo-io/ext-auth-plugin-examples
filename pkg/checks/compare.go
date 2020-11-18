package checks

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


