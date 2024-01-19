package rpc

import "github.com/blang/semver/v4"

var dreamsV = semver.MustParse("0.11.0-dev.17")

// Get current package version
func Version() semver.Version {
	return dreamsV
}
