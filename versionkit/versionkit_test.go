package versionkit

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParseSemanticVersion(t *testing.T) {
	t.Run("basic_version", func(t *testing.T) {
		input := "1.2.3"

		result, err := ParseSemanticVersion(input)

		assert.NoError(t, err)
		assert.Equal(t, uint(1), result.MajorVersion)
		assert.Equal(t, uint(2), result.MinorVersion)
		assert.Equal(t, uint(3), result.PatchVersion)
		assert.Empty(t, result.PreReleaseVersion)
		assert.Empty(t, result.BuildMetadata)
	})

	t.Run("version_with_prerelease", func(t *testing.T) {
		input := "1.2.3-alpha.1"

		result, err := ParseSemanticVersion(input)

		assert.NoError(t, err)
		assert.Equal(t, uint(1), result.MajorVersion)
		assert.Equal(t, uint(2), result.MinorVersion)
		assert.Equal(t, uint(3), result.PatchVersion)
		assert.Equal(t, "alpha.1", result.PreReleaseVersion)
		assert.Empty(t, result.BuildMetadata)
	})

	t.Run("version_with_build_metadata", func(t *testing.T) {
		input := "1.2.3+build.456"

		result, err := ParseSemanticVersion(input)

		assert.NoError(t, err)
		assert.Equal(t, uint(1), result.MajorVersion)
		assert.Equal(t, uint(2), result.MinorVersion)
		assert.Equal(t, uint(3), result.PatchVersion)
		assert.Empty(t, result.PreReleaseVersion)
		assert.Equal(t, "build.456", result.BuildMetadata)
	})

	t.Run("version_with_prerelease_and_build", func(t *testing.T) {
		input := "1.2.3-beta.2+build.789"

		result, err := ParseSemanticVersion(input)

		assert.NoError(t, err)
		assert.Equal(t, uint(1), result.MajorVersion)
		assert.Equal(t, uint(2), result.MinorVersion)
		assert.Equal(t, uint(3), result.PatchVersion)
		assert.Equal(t, "beta.2", result.PreReleaseVersion)
		assert.Equal(t, "build.789", result.BuildMetadata)
	})

	t.Run("zero_version", func(t *testing.T) {
		input := "0.0.0"

		result, err := ParseSemanticVersion(input)

		assert.NoError(t, err)
		assert.Equal(t, uint(0), result.MajorVersion)
		assert.Equal(t, uint(0), result.MinorVersion)
		assert.Equal(t, uint(0), result.PatchVersion)
		assert.Empty(t, result.PreReleaseVersion)
		assert.Empty(t, result.BuildMetadata)
	})

	t.Run("empty_string", func(t *testing.T) {
		input := ""

		_, err := ParseSemanticVersion(input)

		assert.Error(t, err)
	})

	t.Run("invalid_format_-_too_few_parts", func(t *testing.T) {
		input := "1.2"

		_, err := ParseSemanticVersion(input)

		assert.Error(t, err)
	})

	t.Run("invalid_format_-_too_many_parts", func(t *testing.T) {
		input := "1.2.3.4"

		_, err := ParseSemanticVersion(input)

		assert.Error(t, err)
	})

	t.Run("non-numeric_major", func(t *testing.T) {
		input := "a.2.3"

		_, err := ParseSemanticVersion(input)

		assert.Error(t, err)
	})

	t.Run("non-numeric_minor", func(t *testing.T) {
		input := "1.b.3"

		_, err := ParseSemanticVersion(input)

		assert.Error(t, err)
	})

	t.Run("non-numeric_patch", func(t *testing.T) {
		input := "1.2.c"

		_, err := ParseSemanticVersion(input)

		assert.Error(t, err)
	})

	t.Run("multiple_+_signs", func(t *testing.T) {
		input := "1.2.3+build1+build2"

		_, err := ParseSemanticVersion(input)

		assert.Error(t, err)
	})
}

func TestSemanticVersionString(t *testing.T) {
	t.Run("basic_version", func(t *testing.T) {
		version := SemanticVersion{
			MajorVersion: uint(1),
			MinorVersion: uint(2),
			PatchVersion: uint(3),
		}

		result := version.String()

		assert.Equal(t, "1.2.3", result)
	})

	t.Run("version_with_prerelease", func(t *testing.T) {
		version := SemanticVersion{
			MajorVersion:      uint(1),
			MinorVersion:      uint(2),
			PatchVersion:      uint(3),
			PreReleaseVersion: "alpha.1",
		}

		result := version.String()

		assert.Equal(t, "1.2.3-alpha.1", result)
	})

	t.Run("version_with_build_metadata", func(t *testing.T) {
		version := SemanticVersion{
			MajorVersion:  uint(1),
			MinorVersion:  uint(2),
			PatchVersion:  uint(3),
			BuildMetadata: "build.456",
		}

		result := version.String()

		assert.Equal(t, "1.2.3+build.456", result)
	})

	t.Run("version_with_prerelease_and_build", func(t *testing.T) {
		version := SemanticVersion{
			MajorVersion:      uint(1),
			MinorVersion:      uint(2),
			PatchVersion:      uint(3),
			PreReleaseVersion: "beta.2",
			BuildMetadata:     "build.789",
		}

		result := version.String()

		assert.Equal(t, "1.2.3-beta.2+build.789", result)
	})
}

func TestSemanticVersionCompare(t *testing.T) {
	t.Run("1.0.0_<_2.0.0", func(t *testing.T) {
		aVer, err := ParseSemanticVersion("1.0.0")
		assert.NoError(t, err)
		bVer, err := ParseSemanticVersion("2.0.0")
		assert.NoError(t, err)

		result := aVer.Compare(*bVer)

		assert.Equal(t, -1, result)
	})

	t.Run("2.0.0_>_1.0.0", func(t *testing.T) {
		aVer, err := ParseSemanticVersion("2.0.0")
		assert.NoError(t, err)
		bVer, err := ParseSemanticVersion("1.0.0")
		assert.NoError(t, err)

		result := aVer.Compare(*bVer)

		assert.Equal(t, 1, result)
	})

	t.Run("1.0.0_==_1.0.0", func(t *testing.T) {
		aVer, err := ParseSemanticVersion("1.0.0")
		assert.NoError(t, err)
		bVer, err := ParseSemanticVersion("1.0.0")
		assert.NoError(t, err)

		result := aVer.Compare(*bVer)

		assert.Equal(t, 0, result)
	})

	t.Run("1.1.0_>_1.0.0", func(t *testing.T) {
		aVer, err := ParseSemanticVersion("1.1.0")
		assert.NoError(t, err)
		bVer, err := ParseSemanticVersion("1.0.0")
		assert.NoError(t, err)

		result := aVer.Compare(*bVer)

		assert.Equal(t, 1, result)
	})

	t.Run("1.0.1_>_1.0.0", func(t *testing.T) {
		aVer, err := ParseSemanticVersion("1.0.1")
		assert.NoError(t, err)
		bVer, err := ParseSemanticVersion("1.0.0")
		assert.NoError(t, err)

		result := aVer.Compare(*bVer)

		assert.Equal(t, 1, result)
	})

	t.Run("1.0.0-alpha_<_1.0.0", func(t *testing.T) {
		aVer, err := ParseSemanticVersion("1.0.0-alpha")
		assert.NoError(t, err)
		bVer, err := ParseSemanticVersion("1.0.0")
		assert.NoError(t, err)

		result := aVer.Compare(*bVer)

		assert.Equal(t, -1, result)
	})

	t.Run("1.0.0_>_1.0.0-alpha", func(t *testing.T) {
		aVer, err := ParseSemanticVersion("1.0.0")
		assert.NoError(t, err)
		bVer, err := ParseSemanticVersion("1.0.0-alpha")
		assert.NoError(t, err)

		result := aVer.Compare(*bVer)

		assert.Equal(t, 1, result)
	})

	t.Run("1.0.0-alpha_<_1.0.0-beta", func(t *testing.T) {
		aVer, err := ParseSemanticVersion("1.0.0-alpha")
		assert.NoError(t, err)
		bVer, err := ParseSemanticVersion("1.0.0-beta")
		assert.NoError(t, err)

		result := aVer.Compare(*bVer)

		assert.Equal(t, -1, result)
	})

	t.Run("1.0.0-beta_>_1.0.0-alpha", func(t *testing.T) {
		aVer, err := ParseSemanticVersion("1.0.0-beta")
		assert.NoError(t, err)
		bVer, err := ParseSemanticVersion("1.0.0-alpha")
		assert.NoError(t, err)

		result := aVer.Compare(*bVer)

		assert.Equal(t, 1, result)
	})

	t.Run("1.0.0-alpha.1_<_1.0.0-alpha.2", func(t *testing.T) {
		aVer, err := ParseSemanticVersion("1.0.0-alpha.1")
		assert.NoError(t, err)
		bVer, err := ParseSemanticVersion("1.0.0-alpha.2")
		assert.NoError(t, err)

		result := aVer.Compare(*bVer)

		assert.Equal(t, -1, result)
	})

	t.Run("1.0.0-alpha.1_<_1.0.0-alpha.beta", func(t *testing.T) {
		aVer, err := ParseSemanticVersion("1.0.0-alpha.1")
		assert.NoError(t, err)
		bVer, err := ParseSemanticVersion("1.0.0-alpha.beta")
		assert.NoError(t, err)

		result := aVer.Compare(*bVer)

		assert.Equal(t, -1, result)
	})

	t.Run("1.0.0-alpha.beta_>_1.0.0-alpha.1", func(t *testing.T) {
		aVer, err := ParseSemanticVersion("1.0.0-alpha.beta")
		assert.NoError(t, err)
		bVer, err := ParseSemanticVersion("1.0.0-alpha.1")
		assert.NoError(t, err)

		result := aVer.Compare(*bVer)

		assert.Equal(t, 1, result)
	})

	t.Run("1.0.0-rc.1_>_1.0.0-beta.11", func(t *testing.T) {
		aVer, err := ParseSemanticVersion("1.0.0-rc.1")
		assert.NoError(t, err)
		bVer, err := ParseSemanticVersion("1.0.0-beta.11")
		assert.NoError(t, err)

		result := aVer.Compare(*bVer)

		assert.Equal(t, 1, result)
	})

	t.Run("1.0.0+build1_==_1.0.0+build2", func(t *testing.T) {
		aVer, err := ParseSemanticVersion("1.0.0+build1")
		assert.NoError(t, err)
		bVer, err := ParseSemanticVersion("1.0.0+build2")
		assert.NoError(t, err)

		result := aVer.Compare(*bVer)

		assert.Equal(t, 0, result)
	})

	t.Run("1.0.0+build_==_1.0.0", func(t *testing.T) {
		aVer, err := ParseSemanticVersion("1.0.0+build")
		assert.NoError(t, err)
		bVer, err := ParseSemanticVersion("1.0.0")
		assert.NoError(t, err)

		result := aVer.Compare(*bVer)

		assert.Equal(t, 0, result)
	})
}
