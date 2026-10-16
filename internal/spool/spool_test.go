package spool

import (
	"io"
	"testing"

	"AnonymousVertex/internal/jsonx"
)

// TestEncodeJSONMatchesJsonx 验证 EncodeJSON 与 jsonx.Marshal 逐字节一致
// （关 HTML 转义 + 去尾换行），保证发往上游的请求体不变。
func TestEncodeJSONMatchesJsonx(t *testing.T) {
	cases := []any{
		map[string]any{"a": float64(1), "b": "x<y>&z"}, // 含 < > & 验证不转义
		map[string]any{"contents": []any{map[string]any{"role": "user", "parts": []any{map[string]any{"text": "你好"}}}}},
		"plain string",
		[]any{float64(1), float64(2), float64(3)},
	}
	for i, v := range cases {
		buf, err := EncodeJSON(v)
		if err != nil {
			t.Fatalf("case %d EncodeJSON: %v", i, err)
		}
		r, _ := buf.Reader()
		got, _ := io.ReadAll(r)
		want, _ := jsonx.Marshal(v)
		if string(got) != string(want) {
			t.Fatalf("case %d 不一致:\n got=%q\nwant=%q", i, got, want)
		}
		_ = buf.Close()
	}
}

// TestBufferMemOnly 验证内存缓冲：写入、读回完整、Len 正确、不落盘。
func TestBufferMemOnly(t *testing.T) {
	if SpilledBytes() != 0 {
		t.Fatal("SpilledBytes 应为 0")
	}
	SetMaxSpillBytes(123) // 不溢出磁盘，调用不应改变行为

	b := New()
	if _, err := b.Write([]byte("hello")); err != nil {
		t.Fatal(err)
	}
	if _, err := b.Write([]byte("0123456789")); err != nil {
		t.Fatal(err)
	}
	if b.Len() != 15 {
		t.Fatalf("Len 应为 15，got %d", b.Len())
	}
	r, err := b.Reader()
	if err != nil {
		t.Fatal(err)
	}
	got, _ := io.ReadAll(r)
	if string(got) != "hello0123456789" {
		t.Fatalf("读回内容错: %q", got)
	}
	if SpilledBytes() != 0 {
		t.Fatal("写入后 SpilledBytes 仍应为 0（不落盘）")
	}
	if err := b.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}
}
