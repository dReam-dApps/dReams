package rpc

import "github.com/blang/semver/v4"

var dreamsV = semver.MustParse("0.11.0-dev.6")

// Get current package version
func Version() semver.Version {
	return dreamsV
}
