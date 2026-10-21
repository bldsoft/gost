package clickhouse

import (
	"regexp"
	"strings"
	"testing"

	sq "github.com/Masterminds/squirrel"
	"github.com/stretchr/testify/require"

	"github.com/bldsoft/gost/log"
)

func TestExporterSearchParse(t *testing.T) {
	eq := func(s string) sq.Sqlizer {
		return sq.Eq{"field": s}
	}
	notEq := func(s string) sq.Sqlizer {
		return sq.Expr("NOT (?)", eq(s))
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
			name:     "escaped or in sequence",
			search:   "1\\||\\|2",
			expected: sq.Or{eq("1|"), eq("|2")},
		},
		{
			name:     "not operator",
			search:   "!response",
			expected: notEq("response"),
		},
		{
			name:   "not with and and quoted terms",
			search: "!'failed to' & !\"cache miss\"",
			expected: sq.And{
				notEq("failed to"),
				notEq("cache miss"),
			},
		},
		{
			name:     "escaped not",
			search:   "\\!response",
			expected: eq("!response"),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			require.Equal(t, tc.expected, new(ClickHouseLogExporter).parseExpr(tc.search, eq))
		})
	}
}

func TestNormalizeLogExpr(t *testing.T) {
	expr, args := new(ClickHouseLogExporter).normalizeLogExpr("msg")
	require.Equal(t, strings.Repeat("replaceRegexpAll(", 6)+"msg"+strings.Repeat(", ?, ?)", 6), expr)
	require.Len(t, args, 12)

	normalize := func(s string) string {
		for i := 0; i < len(args); i += 2 {
			s = regexp.MustCompile(args[i].(string)).ReplaceAllString(s, args[i+1].(string))
		}
		return s
	}
	tests := []struct {
		name string
		in   string
		want string
	}{
		{name: "quoted", in: `Put "http://host/a/b": context canceled`, want: `Put "…": context canceled`},
		{name: "uuid", in: "user 0f8fad5b-d9cb-469f-a165-70867728950e not found", want: "user <uuid> not found"},
		{name: "ip with port", in: "dial tcp 10.0.0.1:8080: connect: connection refused", want: "dial tcp <ip>: connect: connection refused"},
		{name: "ip", in: "client 192.168.1.10 banned", want: "client <ip> banned"},
		{name: "hex", in: "recorder 6a3a5da6020ea5f52e3826a1: failed to find pool", want: "recorder <hex>: failed to find pool"},
		{name: "0x hex", in: "panic({0x1400012e0c0?, 0x10?})\n\tarchiver.go:212 +0x1d4", want: "panic({<hex>?, <hex>?})\n\tarchiver.go:<N> +<hex>"},
		{name: "0x inside a word", in: "10x5", want: "<N>x<N>"},
		{name: "numbers", in: "Discovery: failed to join cluster: 3 errors occurred", want: "Discovery: failed to join cluster: <N> errors occurred"},
		{name: "letters kept", in: "RESPONSE", want: "RESPONSE"},
		{name: "nothing to mask", in: "unexpected EOF", want: "unexpected EOF"},
		{name: "empty", in: "", want: ""},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			require.Equal(t, tc.want, normalize(tc.in))
			require.Equal(t, tc.want, log.Mask(tc.in))
		})
	}
}
