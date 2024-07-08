package clickhouse

import (
	"testing"

	sq "github.com/Masterminds/squirrel"
	"github.com/stretchr/testify/require"
)

func TestExporterSearchParse(t *testing.T) {
	eq := func(s string) sq.Sqlizer {
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
			expected: eq("1"),
		},
		{
			name:     "or operator",
			search:   "1|2|3",
			expected: sq.Or{eq("1"), eq("2"), eq("3")},
		},
		{
			name:     "or empty",
			search:   "1|",
			expected: eq("1"),
		},
		{
			name:     "and operator",
			search:   "1&2&3",
			expected: sq.And{eq("1"), eq("2"), eq("3")},
		},
		{
			name:     "and empty",
			search:   "1&",
			expected: eq("1"),
		},
		{
			name:   "and/or operator",
			search: "1|2&3|4&5&6|7|8",
			expected: sq.Or{
				eq("1"),
				sq.And{eq("2"), eq("3")},
				sq.And{eq("4"), eq("5"), eq("6")},
				eq("7"),
				eq("8"),
			},
		},
		{
			name:     "escaped and",
			search:   "1\\&2",
			expected: eq("1&2"),
		},
		{
			name:     "escaped or",
			search:   "1\\|2",
			expected: eq("1|2"),
		},
		{
			name:     "escaped or",
			search:   "1\\||\\|2",
			expected: sq.Or{eq("1|"), eq("|2")},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			require.Equal(t, tc.expected, new(ClickHouseLogExporter).parseExpr(tc.search, eq))
		})
	}

}
