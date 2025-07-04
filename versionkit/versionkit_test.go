package versionkit

import (
	"testing"
)

func TestParseSemanticVersion(t *testing.T) {
	tests := []struct {
		name              string
		input             string
		expectedMajor     uint
		expectedMinor     uint
		expectedPatch     uint
		expectedPreRelease string
		expectedBuild     string
		expectError       bool
	}{
		{
			name:              "basic version",
			input:             "1.2.3",
			expectedMajor:     1,
			expectedMinor:     2,
			expectedPatch:     3,
			expectedPreRelease: "",
			expectedBuild:     "",
			expectError:       false,
		},
		{
			name:              "version with pre-release",
			input:             "1.2.3-alpha",
			expectedMajor:     1,
			expectedMinor:     2,
			expectedPatch:     3,
			expectedPreRelease: "alpha",
			expectedBuild:     "",
			expectError:       false,
		},
		{
			name:              "version with build metadata",
			input:             "1.2.3+build.1",
			expectedMajor:     1,
			expectedMinor:     2,
			expectedPatch:     3,
			expectedPreRelease: "",
			expectedBuild:     "build.1",
			expectError:       false,
		},
		{
			name:              "version with pre-release and build metadata",
			input:             "1.2.3-alpha.1+build.1",
			expectedMajor:     1,
			expectedMinor:     2,
			expectedPatch:     3,
			expectedPreRelease: "alpha.1",
			expectedBuild:     "build.1",
			expectError:       false,
		},
		{
			name:              "version with complex pre-release",
			input:             "1.2.3-alpha.beta.1",
			expectedMajor:     1,
			expectedMinor:     2,
			expectedPatch:     3,
			expectedPreRelease: "alpha.beta.1",
			expectedBuild:     "",
			expectError:       false,
		},
		{
			name:              "version with dots in build metadata",
			input:             "1.2.3+build.1.2.3",
			expectedMajor:     1,
			expectedMinor:     2,
			expectedPatch:     3,
			expectedPreRelease: "",
			expectedBuild:     "build.1.2.3",
			expectError:       false,
		},
		{
			name:              "version with complex pre-release and build",
			input:             "1.2.3-alpha.1.2+build.1.2.3",
			expectedMajor:     1,
			expectedMinor:     2,
			expectedPatch:     3,
			expectedPreRelease: "alpha.1.2",
			expectedBuild:     "build.1.2.3",
			expectError:       false,
		},
		{
			name:              "version with zero values",
			input:             "0.0.0",
			expectedMajor:     0,
			expectedMinor:     0,
			expectedPatch:     0,
			expectedPreRelease: "",
			expectedBuild:     "",
			expectError:       false,
		},
		{
			name:              "large version numbers",
			input:             "999.888.777",
			expectedMajor:     999,
			expectedMinor:     888,
			expectedPatch:     777,
			expectedPreRelease: "",
			expectedBuild:     "",
			expectError:       false,
		},
		{
			name:        "empty string",
			input:       "",
			expectError: true,
		},
		{
			name:        "only major version",
			input:       "1",
			expectError: true,
		},
		{
			name:        "only major and minor version",
			input:       "1.2",
			expectError: true,
		},
		{
			name:        "too many version parts",
			input:       "1.2.3.4",
			expectError: true,
		},
		{
			name:        "non-numeric major version",
			input:       "a.2.3",
			expectError: true,
		},
		{
			name:        "non-numeric minor version",
			input:       "1.b.3",
			expectError: true,
		},
		{
			name:        "non-numeric patch version",
			input:       "1.2.c",
			expectError: true,
		},
		{
			name:        "multiple plus signs",
			input:       "1.2.3+build1+build2",
			expectError: true,
		},
		{
			name:        "negative version numbers",
			input:       "-1.2.3",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ParseSemanticVersion(tt.input)

			if tt.expectError {
				if err == nil {
					t.Error("ParseSemanticVersion() expected error but got nil")
				}
				return
			}

			if err != nil {
				t.Errorf("ParseSemanticVersion() unexpected error: %v", err)
				return
			}

			if result.MajorVersion != tt.expectedMajor {
				t.Errorf("ParseSemanticVersion() MajorVersion = %v, want %v", result.MajorVersion, tt.expectedMajor)
			}
			if result.MinorVersion != tt.expectedMinor {
				t.Errorf("ParseSemanticVersion() MinorVersion = %v, want %v", result.MinorVersion, tt.expectedMinor)
			}
			if result.PathVersion != tt.expectedPatch {
				t.Errorf("ParseSemanticVersion() PathVersion = %v, want %v", result.PathVersion, tt.expectedPatch)
			}
			if result.PreReleaseVersion != tt.expectedPreRelease {
				t.Errorf("ParseSemanticVersion() PreReleaseVersion = %q, want %q", result.PreReleaseVersion, tt.expectedPreRelease)
			}
			if result.BuildMetadata != tt.expectedBuild {
				t.Errorf("ParseSemanticVersion() BuildMetadata = %q, want %q", result.BuildMetadata, tt.expectedBuild)
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
				PathVersion:  3,
			},
			expected: "1.2.3",
		},
		{
			name: "version with pre-release",
			version: SemanticVersion{
				MajorVersion:      1,
				MinorVersion:      2,
				PathVersion:       3,
				PreReleaseVersion: "alpha",
			},
			expected: "1.2.3-alpha",
		},
		{
			name: "version with build metadata",
			version: SemanticVersion{
				MajorVersion:  1,
				MinorVersion:  2,
				PathVersion:   3,
				BuildMetadata: "build.1",
			},
			expected: "1.2.3+build.1",
		},
		{
			name: "version with pre-release and build metadata",
			version: SemanticVersion{
				MajorVersion:      1,
				MinorVersion:      2,
				PathVersion:       3,
				PreReleaseVersion: "alpha.1",
				BuildMetadata:     "build.1",
			},
			expected: "1.2.3-alpha.1+build.1",
		},
		{
			name: "zero version",
			version: SemanticVersion{
				MajorVersion: 0,
				MinorVersion: 0,
				PathVersion:  0,
			},
			expected: "0.0.0",
		},
		{
			name: "large version numbers",
			version: SemanticVersion{
				MajorVersion: 999,
				MinorVersion: 888,
				PathVersion:  777,
			},
			expected: "999.888.777",
		},
		{
			name: "complex pre-release",
			version: SemanticVersion{
				MajorVersion:      2,
				MinorVersion:      0,
				PathVersion:       0,
				PreReleaseVersion: "rc.1.2.3",
			},
			expected: "2.0.0-rc.1.2.3",
		},
		{
			name: "complex build metadata",
			version: SemanticVersion{
				MajorVersion:  1,
				MinorVersion:  0,
				PathVersion:   0,
				BuildMetadata: "20130313144700",
			},
			expected: "1.0.0+20130313144700",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.version.String()
			if result != tt.expected {
				t.Errorf("SemanticVersion.String() = %q, want %q", result, tt.expected)
			}
		})
	}
}

func TestParseAndStringRoundTrip(t *testing.T) {
	versions := []string{
		"1.2.3",
		"0.0.1",
		"10.20.30",
		"1.2.3-alpha",
		"1.2.3-alpha.1",
		"1.2.3-alpha.beta",
		"1.2.3+build.1",
		"1.2.3+20130313144700",
		"1.2.3-alpha+build.1",
		"1.2.3-alpha.1+build.1.2.3",
		"1.2.3-rc.1.2.3+build.20230101.123456",
	}

	for _, version := range versions {
		t.Run(version, func(t *testing.T) {
			parsed, err := ParseSemanticVersion(version)
			if err != nil {
				t.Fatalf("ParseSemanticVersion() error = %v", err)
			}

			stringified := parsed.String()
			if stringified != version {
				t.Errorf("Round trip failed: %q -> %q", version, stringified)
			}
		})
	}
}