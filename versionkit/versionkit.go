package versionkit

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

type SemanticVersion struct {
	MajorVersion      uint
	MinorVersion      uint
	PathVersion       uint
	PreReleaseVersion string
	BuildMetadata     string
}

func ParseSemanticVersion(v string) (*SemanticVersion, error) {
	if len(v) == 0 {
		return nil, errors.New("value is empty")
	}

	// First, extract build metadata (everything after '+')
	buildMetadataRaw := ""
	versionWithoutBuild := v
	if plusIndex := strings.Index(v, "+"); plusIndex != -1 {
		if strings.Count(v, "+") > 1 {
			return nil, fmt.Errorf("value %s has more than one + sign", v)
		}
		buildMetadataRaw = v[plusIndex+1:]
		versionWithoutBuild = v[:plusIndex]
	}

	// Next, extract pre-release (everything after '-' in the remaining string)
	preReleaseVersionRaw := ""
	versionCore := versionWithoutBuild
	if dashIndex := strings.Index(versionWithoutBuild, "-"); dashIndex != -1 {
		preReleaseVersionRaw = versionWithoutBuild[dashIndex+1:]
		versionCore = versionWithoutBuild[:dashIndex]
	}

	// Now split the core version (major.minor.patch) on dots
	parts := strings.Split(versionCore, ".")
	if len(parts) != 3 {
		return nil, fmt.Errorf("value %s did not contain major, minor, and patch versions", v)
	}

	majorVersionRaw := parts[0]
	minorVersionRaw := parts[1]
	patchVersionRaw := parts[2]

	sv := &SemanticVersion{}

	// TODO: handle leading 0 in version numbers

	majorVersion, err := strconv.ParseUint(majorVersionRaw, 10, 32)
	if err == nil {
		sv.MajorVersion = uint(majorVersion)
	} else {
		return nil, fmt.Errorf("value %s major version is not numeric", v)
	}

	minorVersion, err := strconv.ParseUint(minorVersionRaw, 10, 32)
	if err == nil {
		sv.MinorVersion = uint(minorVersion)
	} else {
		return nil, fmt.Errorf("value %s major version is not numeric", v)
	}

	patchVersion, err := strconv.ParseUint(patchVersionRaw, 10, 32)
	if err == nil {
		sv.PathVersion = uint(patchVersion)
	} else {
		return nil, fmt.Errorf("value %s major version is not numeric", v)
	}

	// TODO: Validate pre-release and build metadata characters
	sv.PreReleaseVersion = preReleaseVersionRaw
	sv.BuildMetadata = buildMetadataRaw

	return sv, nil
}

func (sv SemanticVersion) String() string {
	core := fmt.Sprintf("%d.%d.%d", sv.MajorVersion, sv.MinorVersion, sv.PathVersion)

	if sv.PreReleaseVersion != "" {
		core = fmt.Sprintf("%s-%s", core, sv.PreReleaseVersion)
	}

	if sv.BuildMetadata != "" {
		core = fmt.Sprintf("%s+%s", core, sv.BuildMetadata)
	}

	return core
}
