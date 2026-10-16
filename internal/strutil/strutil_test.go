package strutil

import (
	"testing"
)

func TestReqID(t *testing.T) {
	id1 := ReqID()
	id2 := ReqID()
	if len(id1) != 24 || len(id2) != 24 {
		t.Fatalf("expected 24 hex characters, got %d and %d", len(id1), len(id2))
	}
	if id1 == id2 {
		t.Fatalf("expected unique IDs, got identical %s", id1)
	}
}

func TestPadBase64(t *testing.T) {
	cases := []struct {
		in   string
		want string
	}{
		{"YWJj", "YWJj"},
		{"YQ", "YQ=="},
		{"YWI", "YWI="},
		{"-_", "+/=="},
	}
	for _, tc := range cases {
		got := PadBase64(tc.in)
		if got != tc.want {
			t.Errorf("PadBase64(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

func TestFirstNonEmpty(t *testing.T) {
	if got := FirstNonEmpty("", "  ", "foo", "bar"); got != "foo" {
		t.Errorf("FirstNonEmpty failed, got %q", got)
	}
	if got := FirstNonEmpty("", "  "); got != "" {
		t.Errorf("FirstNonEmpty on whitespace should return empty, got %q", got)
	}
}

func TestSnakeCamel(t *testing.T) {
	if got := SnakeToCamel("max_output_tokens"); got != "maxOutputTokens" {
		t.Errorf("SnakeToCamel failed: %q", got)
	}
	if got := SnakeToCamel("temperature"); got != "temperature" {
		t.Errorf("SnakeToCamel without underscore should remain untouched: %q", got)
	}
	if got := CamelToSnake("maxOutputTokens"); got != "max_output_tokens" {
		t.Errorf("CamelToSnake failed: %q", got)
	}
}
