// Package version holds the build-time version string for the goulinette
// binary. The default value is overridden at release time:
//
// go build -ldflags "-X github.com/YeiyoNathnael/goulinette/internal/version.Current=1.2.3"
package version

// Current is the goulinette release string. It is set to "dev" by default
// and overridden by the build system via -ldflags at release time.
var Current = "dev"
