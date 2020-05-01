module github.com/solo-io/ext-auth-plugin-examples

go 1.14

require (
	github.com/solo-io/foo v0.0.0-20180207000608-0eeff89b0690
)

replace (
	github.com/solo-io/bar => github.com/solo-io/baz v1.20.4
)
