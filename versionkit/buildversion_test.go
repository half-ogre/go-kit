package versionkit

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetBuildInfo(t *testing.T) {
	t.Run("returns_a_build_info_struct", func(t *testing.T) {
		result := GetBuildInfo()

		assert.NotNil(t, result)
	})

	t.Run("returns_build_info_populated_from_runtime", func(t *testing.T) {
		result := GetBuildInfo()

		// We can't assert exact values since they depend on how the test is built,
		// but we can verify the struct is returned
		assert.NotNil(t, result)
		// Version might be a pseudo-version or empty, both are valid
		// GitCommit might be set if built with VCS info
		// BuildDate might be set if built with VCS info
	})
}

func TestGetBuildVersion(t *testing.T) {
	t.Run("returns_the_version_when_set", func(t *testing.T) {
		bi := &BuildInfo{
			Version:   "theVersion",
			GitCommit: "aCommit",
			BuildDate: "aDate",
		}

		result := bi.GetBuildVersion()

		assert.Equal(t, "theVersion", result)
	})

	t.Run("returns_dev_when_version_is_empty", func(t *testing.T) {
		bi := &BuildInfo{
			Version:   "",
			GitCommit: "aCommit",
			BuildDate: "aDate",
		}

		result := bi.GetBuildVersion()

		assert.Equal(t, "dev", result)
	})

	t.Run("returns_dev_when_version_is_not_set", func(t *testing.T) {
		bi := &BuildInfo{}

		result := bi.GetBuildVersion()

		assert.Equal(t, "dev", result)
	})
}

func TestGetBuildCommit(t *testing.T) {
	t.Run("returns_the_commit_when_set", func(t *testing.T) {
		bi := &BuildInfo{
			Version:   "aVersion",
			GitCommit: "theCommit",
			BuildDate: "aDate",
		}

		result := bi.GetBuildCommit()

		assert.Equal(t, "theCommit", result)
	})

	t.Run("returns_unknown_when_commit_is_empty", func(t *testing.T) {
		bi := &BuildInfo{
			Version:   "aVersion",
			GitCommit: "",
			BuildDate: "aDate",
		}

		result := bi.GetBuildCommit()

		assert.Equal(t, "unknown", result)
	})

	t.Run("returns_unknown_when_commit_is_not_set", func(t *testing.T) {
		bi := &BuildInfo{}

		result := bi.GetBuildCommit()

		assert.Equal(t, "unknown", result)
	})
}

func TestGetBuildDate(t *testing.T) {
	t.Run("returns_the_date_when_set", func(t *testing.T) {
		bi := &BuildInfo{
			Version:   "aVersion",
			GitCommit: "aCommit",
			BuildDate: "theDate",
		}

		result := bi.GetBuildDate()

		assert.Equal(t, "theDate", result)
	})

	t.Run("returns_unknown_when_date_is_empty", func(t *testing.T) {
		bi := &BuildInfo{
			Version:   "aVersion",
			GitCommit: "aCommit",
			BuildDate: "",
		}

		result := bi.GetBuildDate()

		assert.Equal(t, "unknown", result)
	})

	t.Run("returns_unknown_when_date_is_not_set", func(t *testing.T) {
		bi := &BuildInfo{}

		result := bi.GetBuildDate()

		assert.Equal(t, "unknown", result)
	})
}

func TestBuildInfoString(t *testing.T) {
	t.Run("formats_all_fields_when_set", func(t *testing.T) {
		bi := &BuildInfo{
			Version:   "theVersion",
			GitCommit: "theCommit",
			BuildDate: "theDate",
		}

		result := bi.String()

		assert.Equal(t, "version theVersion (commit: theCommit, built: theDate)", result)
	})

	t.Run("uses_dev_for_missing_version", func(t *testing.T) {
		bi := &BuildInfo{
			Version:   "",
			GitCommit: "aCommit",
			BuildDate: "aDate",
		}

		result := bi.String()

		assert.Equal(t, "version dev (commit: aCommit, built: aDate)", result)
	})

	t.Run("uses_unknown_for_missing_commit", func(t *testing.T) {
		bi := &BuildInfo{
			Version:   "aVersion",
			GitCommit: "",
			BuildDate: "aDate",
		}

		result := bi.String()

		assert.Equal(t, "version aVersion (commit: unknown, built: aDate)", result)
	})

	t.Run("uses_unknown_for_missing_date", func(t *testing.T) {
		bi := &BuildInfo{
			Version:   "aVersion",
			GitCommit: "aCommit",
			BuildDate: "",
		}

		result := bi.String()

		assert.Equal(t, "version aVersion (commit: aCommit, built: unknown)", result)
	})

	t.Run("uses_defaults_for_all_missing_fields", func(t *testing.T) {
		bi := &BuildInfo{}

		result := bi.String()

		assert.Equal(t, "version dev (commit: unknown, built: unknown)", result)
	})
}
