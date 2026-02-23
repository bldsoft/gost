package clickhouse

import (
	"testing"

	sq "github.com/Masterminds/squirrel"
	"github.com/stretchr/testify/require"
)

func TestExporterSearchParse(t *testing.T) {
	eq := func(s string, not bool) sq.Sqlizer {
		if not {
			return sq.Eq{"field": "!" + s}
		}
		return sq.Eq{"field": s}
	}
	tests := []struct {
		name     string
		search   string
		expected sq.Sqlizer
	}{
		{
			name:     "no operators",
			search:   "1",
			expected: eq("1", false),
		},
		{
			name:     "or operator",
			search:   "1|2|3",
			expected: sq.Or{eq("1", false), eq("2", false), eq("3", false)},
		},
		{
			name:     "or empty",
			search:   "1|",
			expected: eq("1", false),
		},
		{
			name:     "and operator",
			search:   "1&2&3",
			expected: sq.And{eq("1", false), eq("2", false), eq("3", false)},
		},
		{
			name:     "and empty",
			search:   "1&",
			expected: eq("1", false),
		},
		{
			name:   "and/or operator",
			search: "1|2&3|4&5&6|7|8",
			expected: sq.Or{
				eq("1", false),
				sq.And{eq("2", false), eq("3", false)},
				sq.And{eq("4", false), eq("5", false), eq("6", false)},
				eq("7", false),
				eq("8", false),
			},
		},
		{
			name:     "escaped and",
			search:   "1\\&2",
			expected: eq("1&2", false),
		},
		{
			name:     "escaped or",
			search:   "1\\|2",
			expected: eq("1|2", false),
		},
		{
			name:     "escaped or in sequence",
			search:   "1\\||\\|2",
			expected: sq.Or{eq("1|", false), eq("|2", false)},
		},
		{
			name:     "not operator",
			search:   "!response",
			expected: eq("response", true),
		},
		{
			name:   "not with and and quoted terms",
			search: "!'failed to' & !\"cache miss\"",
			expected: sq.And{
				eq("failed to", true),
				eq("cache miss", true),
			},
		},
		{
			name:     "escaped not",
			search:   "\\!response",
			expected: eq("!response", false),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			require.Equal(t, tc.expected, new(ClickHouseLogExporter).parseExpr(tc.search, eq))
		})
	}
}
