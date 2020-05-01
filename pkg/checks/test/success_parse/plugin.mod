module github.com/solo-io/ext-auth-plugin-examples

go 1.14

require (
	github.com/solo-io/foo v0.0.0-20180207000608-0eeff89b0690
    // some comment...
	github.com/solo-io/baz v1.2.5

	github.com/solo-io/bar v1.2.3 //indirect
)

replace (
    //Dependency with version replaced
	github.com/solo-io/bar v1.2.3 => github.com/solo-io/bar v1.2.4

    // Dependency with version and name replaced
	github.com/solo-io/baz v1.2.5 => github.com/solo-io/barfoo v1.2.4

    // Dependency with no version
	github.com/solo-io/foo => github.com/solo-io/bar v1.2.4
)


