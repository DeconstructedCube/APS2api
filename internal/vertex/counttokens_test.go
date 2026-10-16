package vertex

import "testing"

func TestEstimateTokens(t *testing.T) {
	cases := []struct {
		name     string
		contents []any
		want     int
	}{
		{
			name:     "nil contents",
			contents: nil,
			want:     0,
		},
		{
			name:     "empty contents",
			contents: []any{},
			want:     0,
		},
		{
			name: "single string content",
			contents: []any{
				"Hello World",
			},
			want: 3, // "Hello World" -> 11 chars * 0.25 = 2.75 -> 3
		},
		{
			name: "single content with ascii text part",
			contents: []any{
				map[string]any{
					"parts": []any{
						map[string]any{"text": "Hello, world!"}, // 13 chars * 0.25 = 3.25 -> 4
					},
				},
			},
			want: 4,
		},
		{
			name: "single content with chinese text part",
			contents: []any{
				map[string]any{
					"parts": []any{
						map[string]any{"text": "你好，世界"}, // 5 chars * 1.5 = 7.5 -> 8
					},
				},
			},
			want: 8,
		},
		{
			name: "mixed ascii and chinese text part",
			contents: []any{
				map[string]any{
					"parts": []any{
						map[string]any{"text": "Hello 世界"}, // 6 ascii (1.5) + 2 chinese (3.0) = 4.5 -> 5
					},
				},
			},
			want: 5,
		},
		{
			name: "Gemini style inlineData image part",
			contents: []any{
				map[string]any{
					"parts": []any{
						map[string]any{
							"inlineData": map[string]any{
								"mimeType": "image/png",
								"data":     "base64data",
							},
						},
					},
				},
			},
			want: 1024,
		},
		{
			name: "Gemini style fileData image part",
			contents: []any{
				map[string]any{
					"parts": []any{
						map[string]any{
							"fileData": map[string]any{
								"mimeType": "image/webp",
								"fileUri":  "gs://bucket/img.webp",
							},
						},
					},
				},
			},
			want: 1024,
		},
		{
			name: "mixed text and image parts",
			contents: []any{
				map[string]any{
					"parts": []any{
						map[string]any{"text": "Analyze this image:"}, // 20 chars * 0.25 = 5
						map[string]any{
							"inlineData": map[string]any{
								"mimeType": "image/jpeg",
								"data":     "xyz",
							},
						},
						map[string]any{"text": "谢谢"}, // 2 * 1.5 = 3
					},
				},
			},
			want: 1032, // 5 + 1024 + 3 = 1032
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := EstimateTokens(c.contents); got != c.want {
				t.Errorf("EstimateTokens(%v) = %d, want %d", c.name, got, c.want)
			}
		})
	}
}
