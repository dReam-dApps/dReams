package rpc

import "github.com/blang/semver/v4"

var dreamsV = semver.MustParse("0.11.1-dev.2")

// Get current package version
func Version() semver.Version {
	return dreamsV
}
