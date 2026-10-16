package api

import (
	"net/http"
	"strings"

	"AnonymousVertex/internal/config"
)

// resolveConfiguredModel 统一处理模型别名与模型启用状态。
func resolveConfiguredModel(rawModel string, cfg config.ConfigProvider) (actualModel string, ok bool) {
	actualModel = cfg.ResolveModelName(strings.TrimSpace(rawModel))
	entry, exists := cfg.LookupModel(actualModel)
	if !exists || !entry.Enabled {
		return actualModel, false
	}
	return actualModel, true
}

func oaiModelNotFound(w http.ResponseWriter, model string) {
	writeJSON(w, http.StatusNotFound, map[string]any{"error": map[string]any{
		"message": "Model '" + model + "' not found.", "type": "invalid_request_error", "code": "model_not_found", "param": "model",
	}})
}

func geminiModelNotFound(w http.ResponseWriter, model string) {
	writeJSON(w, http.StatusNotFound, map[string]any{"error": map[string]any{
		"code": 404, "message": "Model '" + model + "' not found.", "status": "NOT_FOUND",
	}})
}

// supportedGenerationMethods 返回模型详情里声明的支持方法。
func supportedGenerationMethods() []any {
	return []any{"generateContent", "streamGenerateContent"}
}

// geminiModelInfo 构造单个 Gemini 模型详情对象。
func geminiModelInfo(name string) map[string]any {
	return map[string]any{
		"name":                       "models/" + name,
		"version":                    name,
		"displayName":                name,
		"description":                "Vertex AI Studio anonymous model",
		"inputTokenLimit":            1048576,
		"outputTokenLimit":           65536,
		"supportedGenerationMethods": supportedGenerationMethods(),
	}
}
