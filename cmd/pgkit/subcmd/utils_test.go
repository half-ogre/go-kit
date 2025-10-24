package subcmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestQuoteIdentifier(t *testing.T) {
	t.Run("quotes_a_simple_identifier", func(t *testing.T) {
		result := quoteIdentifier("mydb")

		assert.Equal(t, `"mydb"`, result)
	})

	t.Run("escapes_double_quotes_in_identifier", func(t *testing.T) {
		result := quoteIdentifier(`my"db`)

		assert.Equal(t, `"my""db"`, result)
	})

	t.Run("escapes_multiple_double_quotes_in_identifier", func(t *testing.T) {
		result := quoteIdentifier(`my"special"db`)

		assert.Equal(t, `"my""special""db"`, result)
	})

	t.Run("handles_empty_identifier", func(t *testing.T) {
		result := quoteIdentifier("")

		assert.Equal(t, `""`, result)
	})
}
