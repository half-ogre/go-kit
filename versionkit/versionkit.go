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
	PatchVersion      uint   // Fixed typo: was PathVersion
	PreReleaseVersion string
	BuildMetadata     string
}

func ParseSemanticVersion(v string) (*SemanticVersion, error) {
	if len(v) == 0 {
		return nil, errors.New("value is empty")
	}

	// First handle build metadata (everything after +)
	buildMetadataRaw := ""
	if strings.Contains(v, "+") {
		buildParts := strings.Split(v, "+")
		if len(buildParts) > 2 {
			return nil, fmt.Errorf("value %s has more than one + sign", v)
		}
		buildMetadataRaw = buildParts[1]
		v = buildParts[0] // Remove build metadata for further processing
	}

	// Then handle pre-release (everything after - but before +)
	preReleaseVersionRaw := ""
	if strings.Contains(v, "-") {
		preParts := strings.SplitN(v, "-", 2)
		preReleaseVersionRaw = preParts[1]
		v = preParts[0] // Remove pre-release for further processing
	}

	// Now split the core version (should be exactly 3 parts)
	parts := strings.Split(v, ".")
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
		return nil, fmt.Errorf("value %s minor version is not numeric", v)
	}

	patchVersion, err := strconv.ParseUint(patchVersionRaw, 10, 32)
	if err == nil {
		sv.PatchVersion = uint(patchVersion)
	} else {
		return nil, fmt.Errorf("value %s patch version is not numeric", v)
	}

	// TODO: Validate pre-release and build metadata characters
	sv.PreReleaseVersion = preReleaseVersionRaw
	sv.BuildMetadata = buildMetadataRaw

	return sv, nil
}

func (sv SemanticVersion) String() string {
	core := fmt.Sprintf("%d.%d.%d", sv.MajorVersion, sv.MinorVersion, sv.PatchVersion)

	if sv.PreReleaseVersion != "" {
		core = fmt.Sprintf("%s-%s", core, sv.PreReleaseVersion)
	}

	if sv.BuildMetadata != "" {
		core = fmt.Sprintf("%s+%s", core, sv.BuildMetadata)
	}

	return core
}

// Compare returns -1 if sv < other, 0 if sv == other, 1 if sv > other
// according to semantic versioning precedence rules
func (sv SemanticVersion) Compare(other SemanticVersion) int {
	// Compare major version
	if sv.MajorVersion < other.MajorVersion {
		return -1
	}
	if sv.MajorVersion > other.MajorVersion {
		return 1
	}

	// Compare minor version
	if sv.MinorVersion < other.MinorVersion {
		return -1
	}
	if sv.MinorVersion > other.MinorVersion {
		return 1
	}

	// Compare patch version
	if sv.PatchVersion < other.PatchVersion {
		return -1
	}
	if sv.PatchVersion > other.PatchVersion {
		return 1
	}

	// Compare pre-release versions
	// A pre-release version has lower precedence than a normal version
	if sv.PreReleaseVersion == "" && other.PreReleaseVersion != "" {
		return 1
	}
	if sv.PreReleaseVersion != "" && other.PreReleaseVersion == "" {
		return -1
	}
	if sv.PreReleaseVersion != "" && other.PreReleaseVersion != "" {
		return comparePrerelease(sv.PreReleaseVersion, other.PreReleaseVersion)
	}

	// Build metadata does not affect version precedence
	return 0
}

// comparePrerelease compares two pre-release version strings
func comparePrerelease(a, b string) int {
	aParts := strings.Split(a, ".")
	bParts := strings.Split(b, ".")

	maxLen := len(aParts)
	if len(bParts) > maxLen {
		maxLen = len(bParts)
	}

	for i := 0; i < maxLen; i++ {
		var aPart, bPart string
		if i < len(aParts) {
			aPart = aParts[i]
		}
		if i < len(bParts) {
			bPart = bParts[i]
		}

		// If one side is missing, it has lower precedence
		if aPart == "" && bPart != "" {
			return -1
		}
		if aPart != "" && bPart == "" {
			return 1
		}

		// Try to parse as numbers
		aNum, aIsNum := parseUintOrZero(aPart)
		bNum, bIsNum := parseUintOrZero(bPart)

		if aIsNum && bIsNum {
			if aNum < bNum {
				return -1
			}
			if aNum > bNum {
				return 1
			}
		} else if aIsNum && !bIsNum {
			// Numeric identifiers always have lower precedence than non-numeric
			return -1
		} else if !aIsNum && bIsNum {
			// Non-numeric identifiers always have higher precedence than numeric
			return 1
		} else {
			// Both are non-numeric, compare lexically
			if aPart < bPart {
				return -1
			}
			if aPart > bPart {
				return 1
			}
		}
	}

	return 0
}

func parseUintOrZero(s string) (uint64, bool) {
	if num, err := strconv.ParseUint(s, 10, 64); err == nil {
		return num, true
	}
	return 0, false
}