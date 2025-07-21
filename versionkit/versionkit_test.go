package versionkit

import (
	"testing"
)

func TestParseSemanticVersion(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expected    *SemanticVersion
		expectError bool
	}{
		{
			name:  "basic version",
			input: "1.2.3",
			expected: &SemanticVersion{
				MajorVersion: 1,
				MinorVersion: 2,
				PatchVersion: 3,
			},
		},
		{
			name:  "version with prerelease",
			input: "1.2.3-alpha.1",
			expected: &SemanticVersion{
				MajorVersion:      1,
				MinorVersion:      2,
				PatchVersion:      3,
				PreReleaseVersion: "alpha.1",
			},
		},
		{
			name:  "version with build metadata",
			input: "1.2.3+build.456",
			expected: &SemanticVersion{
				MajorVersion:  1,
				MinorVersion:  2,
				PatchVersion:  3,
				BuildMetadata: "build.456",
			},
		},
		{
			name:  "version with prerelease and build",
			input: "1.2.3-beta.2+build.789",
			expected: &SemanticVersion{
				MajorVersion:      1,
				MinorVersion:      2,
				PatchVersion:      3,
				PreReleaseVersion: "beta.2",
				BuildMetadata:     "build.789",
			},
		},
		{
			name:  "zero version",
			input: "0.0.0",
			expected: &SemanticVersion{
				MajorVersion: 0,
				MinorVersion: 0,
				PatchVersion: 0,
			},
		},
		{
			name:        "empty string",
			input:       "",
			expectError: true,
		},
		{
			name:        "invalid format - too few parts",
			input:       "1.2",
			expectError: true,
		},
		{
			name:        "invalid format - too many parts",
			input:       "1.2.3.4",
			expectError: true,
		},
		{
			name:        "non-numeric major",
			input:       "a.2.3",
			expectError: true,
		},
		{
			name:        "non-numeric minor",
			input:       "1.b.3",
			expectError: true,
		},
		{
			name:        "non-numeric patch",
			input:       "1.2.c",
			expectError: true,
		},
		{
			name:        "multiple + signs",
			input:       "1.2.3+build1+build2",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseSemanticVersion(tt.input)

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
				return
			}

			if result.MajorVersion != tt.expected.MajorVersion {
				t.Errorf("MajorVersion = %d, want %d", result.MajorVersion, tt.expected.MajorVersion)
			}
			if result.MinorVersion != tt.expected.MinorVersion {
				t.Errorf("MinorVersion = %d, want %d", result.MinorVersion, tt.expected.MinorVersion)
			}
			if result.PatchVersion != tt.expected.PatchVersion {
				t.Errorf("PatchVersion = %d, want %d", result.PatchVersion, tt.expected.PatchVersion)
			}
			if result.PreReleaseVersion != tt.expected.PreReleaseVersion {
				t.Errorf("PreReleaseVersion = %q, want %q", result.PreReleaseVersion, tt.expected.PreReleaseVersion)
			}
			if result.BuildMetadata != tt.expected.BuildMetadata {
				t.Errorf("BuildMetadata = %q, want %q", result.BuildMetadata, tt.expected.BuildMetadata)
			}
		})
	}
}

func TestSemanticVersionString(t *testing.T) {
	tests := []struct {
		name     string
		version  SemanticVersion
		expected string
	}{
		{
			name: "basic version",
			version: SemanticVersion{
				MajorVersion: 1,
				MinorVersion: 2,
				PatchVersion: 3,
			},
			expected: "1.2.3",
		},
		{
			name: "version with prerelease",
			version: SemanticVersion{
				MajorVersion:      1,
				MinorVersion:      2,
				PatchVersion:      3,
				PreReleaseVersion: "alpha.1",
			},
			expected: "1.2.3-alpha.1",
		},
		{
			name: "version with build metadata",
			version: SemanticVersion{
				MajorVersion:  1,
				MinorVersion:  2,
				PatchVersion:  3,
				BuildMetadata: "build.456",
			},
			expected: "1.2.3+build.456",
		},
		{
			name: "version with prerelease and build",
			version: SemanticVersion{
				MajorVersion:      1,
				MinorVersion:      2,
				PatchVersion:      3,
				PreReleaseVersion: "beta.2",
				BuildMetadata:     "build.789",
			},
			expected: "1.2.3-beta.2+build.789",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.version.String()
			if result != tt.expected {
				t.Errorf("String() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestSemanticVersionCompare(t *testing.T) {
	tests := []struct {
		name     string
		a        string
		b        string
		expected int
	}{
		// Basic version comparisons
		{"1.0.0 < 2.0.0", "1.0.0", "2.0.0", -1},
		{"2.0.0 > 1.0.0", "2.0.0", "1.0.0", 1},
		{"1.0.0 == 1.0.0", "1.0.0", "1.0.0", 0},
		{"1.1.0 > 1.0.0", "1.1.0", "1.0.0", 1},
		{"1.0.1 > 1.0.0", "1.0.1", "1.0.0", 1},

		// Pre-release comparisons
		{"1.0.0-alpha < 1.0.0", "1.0.0-alpha", "1.0.0", -1},
		{"1.0.0 > 1.0.0-alpha", "1.0.0", "1.0.0-alpha", 1},
		{"1.0.0-alpha < 1.0.0-beta", "1.0.0-alpha", "1.0.0-beta", -1},
		{"1.0.0-beta > 1.0.0-alpha", "1.0.0-beta", "1.0.0-alpha", 1},
		{"1.0.0-alpha.1 < 1.0.0-alpha.2", "1.0.0-alpha.1", "1.0.0-alpha.2", -1},
		{"1.0.0-alpha.1 < 1.0.0-alpha.beta", "1.0.0-alpha.1", "1.0.0-alpha.beta", -1},
		{"1.0.0-alpha.beta > 1.0.0-alpha.1", "1.0.0-alpha.beta", "1.0.0-alpha.1", 1},
		{"1.0.0-rc.1 > 1.0.0-beta.11", "1.0.0-rc.1", "1.0.0-beta.11", 1},

		// Build metadata should not affect comparison
		{"1.0.0+build1 == 1.0.0+build2", "1.0.0+build1", "1.0.0+build2", 0},
		{"1.0.0+build == 1.0.0", "1.0.0+build", "1.0.0", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			aVer, err := ParseSemanticVersion(tt.a)
			if err != nil {
				t.Fatalf("Failed to parse version A: %v", err)
			}

			bVer, err := ParseSemanticVersion(tt.b)
			if err != nil {
				t.Fatalf("Failed to parse version B: %v", err)
			}

			result := aVer.Compare(*bVer)
			if result != tt.expected {
				t.Errorf("Compare(%s, %s) = %d, want %d", tt.a, tt.b, result, tt.expected)
			}
		})
	}
}