package kit

import (
	"regexp"
	"strings"
)

// GenerateSlug creates a URL-safe slug from a string
func GenerateSlug(s string) string {
	// Convert to lowercase
	slug := strings.ToLower(s)
	// Replace spaces with hyphens
	slug = strings.ReplaceAll(slug, " ", "-")
	// Remove all non-alphanumeric characters except hyphens
	reg := regexp.MustCompile(`[^a-z0-9-]+`)
	slug = reg.ReplaceAllString(slug, "")
	// Remove duplicate hyphens
	reg = regexp.MustCompile(`-+`)
	slug = reg.ReplaceAllString(slug, "-")
	// Trim hyphens from start and end
	slug = strings.Trim(slug, "-")
	return slug
}
