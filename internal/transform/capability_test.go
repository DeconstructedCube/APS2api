package transform

import "testing"

func TestApplyModelCapabilitiesKnownModelClampsExplicitValues(t *testing.T) {
	payload := map[string]any{"generationConfig": map[string]any{"temperature": 9.0, "topP": -1.0, "maxOutputTokens": float64(999999)}}
	ApplyModelCapabilities(payload, "gemini-3.5-flash")
	g := payload["generationConfig"].(map[string]any)
	if g["temperature"] != 2.0 || g["topP"] != 0.0 || g["maxOutputTokens"] != 65536 {
		t.Fatalf("clamped generation config: %#v", g)
	}
}

func TestApplyModelCapabilitiesUnknownModelPassesThrough(t *testing.T) {
	payload := map[string]any{"generationConfig": map[string]any{"temperature": 9.0, "topP": -1.0, "maxOutputTokens": float64(999999)}}
	ApplyModelCapabilities(payload, "future-model")
	g := payload["generationConfig"].(map[string]any)
	if g["temperature"] != 9.0 || g["topP"] != -1.0 || g["maxOutputTokens"] != float64(999999) {
		t.Fatalf("unknown model was modified: %#v", g)
	}
}
func TestClampIntRejectsNonIntegerFloat(t *testing.T) {
	if got := clampInt(1.9, 1, 100); got != 1.9 {
		t.Fatalf("clampInt should not truncate float: got %v", got)
	}
	if got := clampInt(5.0, 1, 100); got != 5 {
		t.Fatalf("clampInt should convert integer float to int: got %v", got)
	}
}

func TestNormalizeBuiltinToolPreservesGoogleSearchConfig(t *testing.T) {
	tool := map[string]any{
		"type":          "google_search",
		"google_search": map[string]any{"threshold": 0.5},
	}
	normalized := normalizeBuiltinTool(tool)
	if normalized == nil {
		t.Fatal("normalizeBuiltinTool returned nil")
	}
	gs, ok := normalized["googleSearch"].(map[string]any)
	if !ok || gs["threshold"] != 0.5 {
		t.Fatalf("googleSearch config was not preserved: %#v", normalized)
	}
}

func TestReleaseStreamCreated(t *testing.T) {
	reqID := "test-release-req"
	_ = streamCreated(reqID)
	if _, ok := streamCreatedByRequest.Load(reqID); !ok {
		t.Fatal("expected request entry in cache")
	}
	ReleaseStreamCreated(reqID)
	if _, ok := streamCreatedByRequest.Load(reqID); ok {
		t.Fatal("expected request entry to be deleted after release")
	}
}
