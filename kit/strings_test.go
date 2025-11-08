package kit

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGenerateSlug(t *testing.T) {
	t.Run("converts_to_lowercase", func(t *testing.T) {
		result := GenerateSlug("Software Development")

		assert.Equal(t, "software-development", result)
	})

	t.Run("replaces_spaces_with_hyphens", func(t *testing.T) {
		result := GenerateSlug("My Test Title")

		assert.Equal(t, "my-test-title", result)
	})

	t.Run("removes_special_characters", func(t *testing.T) {
		result := GenerateSlug("Title! With@ Special# Characters$")

		assert.Equal(t, "title-with-special-characters", result)
	})

	t.Run("removes_duplicate_hyphens", func(t *testing.T) {
		result := GenerateSlug("Title   With   Spaces")

		assert.Equal(t, "title-with-spaces", result)
	})

	t.Run("trims_hyphens_from_start_and_end", func(t *testing.T) {
		result := GenerateSlug("  Title  ")

		assert.Equal(t, "title", result)
	})

	t.Run("handles_empty_string", func(t *testing.T) {
		result := GenerateSlug("")

		assert.Equal(t, "", result)
	})

	t.Run("handles_only_special_characters", func(t *testing.T) {
		result := GenerateSlug("!@#$%")

		assert.Equal(t, "", result)
	})
}
