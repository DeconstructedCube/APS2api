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

func TestExtractProxyURIName(t *testing.T) {
	if got := ExtractProxyURIName("vmess://eyJwcyI6ImxlZ2FjeS1uYW1lIn0="); got != "legacy-name" {
		t.Fatalf("unexpected vmess ps name %q", got)
	}
	if got := ExtractProxyURIName("socks5://127.0.0.1:1080#%E5%85%A5%E5%8F%A3"); got != "入口" {
		t.Fatalf("unexpected fragment name %q", got)
	}
}

