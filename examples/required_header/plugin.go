package main

import (
	"github.com/solo-io/ext-auth-plugins/api"
	impl "github.com/solo-io/ext-auth-plugins/examples/required_header/pkg"
)

func main() {}

// Compile-time assertion
var _ api.ExtAuthPlugin = new(impl.RequiredHeaderPlugin)

// This is the exported symbol that GlooE will look for.
//noinspection GoUnusedGlobalVariable
var Plugin impl.RequiredHeaderPlugin
