package main

import (
	"fmt"
	"os"

	"github.com/solo-io/go-utils/versionutils"
)

var (
	lastGoPathGlooEVersions = []versionutils.Version{
		{
			Major: 1,
			Minor: 3,
			Patch: 3,
		},
		{
			Major:        1,
			Minor:        4,
			Patch:        0,
			Label:        "beta",
			LabelVersion: 2,
		},
	}
)

// script that outputs to stdout the go build mode of GlooE based on the version (prefixed with a v) being built
// "gomod" for newer versions of GlooE
// "gopath" for older versions of GlooE
func main() {
	if len(os.Args) != 2 {
		fmt.Printf("Must provide 1 arguments: \n\t- GlooE version (must be prefixed with 'v')\n")
		os.Exit(1)
	}

	version, err := versionutils.ParseVersion(os.Args[1])
	if err != nil {
		panic(err)
	}

	defaultBuildMode := "gopath"

	for _, v := range lastGoPathGlooEVersions {
		isGreater, determinable := version.IsGreaterThan(v)
		if isGreater && determinable {
			defaultBuildMode = "gomod"
			if v.Major == version.Major && v.Minor == version.Minor {
				fmt.Println("gomod")
				return
			}
		} else if determinable {
			if v.Major == version.Major && v.Minor == version.Minor {
				fmt.Println("gopath")
				return
			}
		}
	}

	fmt.Println(defaultBuildMode)
}
