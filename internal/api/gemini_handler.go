package api

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"AnonymousVertex/internal/jsonx"
	"AnonymousVertex/internal/vertex"
)

type GeminiHandler struct {
	handler
}

func (g *GeminiHandler) handleModelsSubtree(w http.ResponseWriter, r *http.Request) {
	if !strings.HasPrefix(r.URL.Path, "/v1beta/models/") {
		g.geminiError(w, http.StatusNotFound, "not found", "NOT_FOUND")
		return
	}
	rest := strings.TrimPrefix(r.URL.Path, "/v1beta/models/")
	if rest == "" {
		g.geminiError(w, http.StatusNotFound, "not found", "NOT_FOUND")
		return
	}

	model := rest
	method := ""
	if idx := strings.LastIndex(rest, ":"); idx != -1 {
		model = rest[:idx]
		method = rest[idx+1:]
	}

	switch method {
	case "":
		if r.Method != http.MethodGet {
			g.geminiError(w, http.StatusMethodNotAllowed, "method not allowed", "INVALID_ARGUMENT")
			return
		}
		g.handleModelInfo(w, model)
	case "generateContent":
		g.requirePost(w, r, func() { g.handleGeminiGenerate(w, r, model) })
	case "streamGenerateContent":
		g.requirePost(w, r, func() { g.handleGeminiStreamGenerate(w, r, model) })
	default:
		g.geminiError(w, http.StatusNotFound, "未知方法 "+method+" (unknown method)", "NOT_FOUND")
	}
}

func (g *GeminiHandler) requirePost(w http.ResponseWriter, r *http.Request, fn func()) {
	if r.Method != http.MethodPost {
		g.geminiError(w, http.StatusMethodNotAllowed, "method not allowed", "INVALID_ARGUMENT")
		return
	}
	fn()
}

func (g *GeminiHandler) readGeminiBody(w http.ResponseWriter, r *http.Request) (map[string]any, bool) {
	var body map[string]any
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		if _, ok := err.(*json.SyntaxError); ok && strings.Contains(err.Error(), "invalid UTF-8") {
			g.geminiError(w, http.StatusBadRequest, "请求体编码错误，需为 UTF-8 (request body must be UTF-8 encoded)", "INVALID_ARGUMENT")
			return nil, false
		}
		g.geminiError(w, http.StatusBadRequest, "请求格式错误，JSON 解析失败 (invalid JSON)", "INVALID_ARGUMENT")
		return nil, false
	}
	if body == nil {
		body = make(map[string]any)
	}
	return body, true
}

// handleGeminiGenerate 零篡改直通生成
func (g *GeminiHandler) handleGeminiGenerate(w http.ResponseWriter, r *http.Request, model string) {
	actualModel, modelOK := resolveConfiguredModel(model, g.cfg)
	if !modelOK {
		geminiModelNotFound(w, model)
		return
	}
	body, ok := g.readGeminiBody(w, r)
	if !ok {
		return
	}
	if reqObj, ok2 := body["generateContentRequest"].(map[string]any); ok2 {
		body = reqObj
	}
	log.Printf("[Server] [GeminiGenerate] 收到请求: 模型=%s, 真模型=%s", model, actualModel)

	resp, vErr := g.vc.CompleteChat(r.Context(), actualModel, body)
	if vErr != nil {
		ve := toVertexError(vErr)
		if isSafetyBlock(ve) {
			writeJSON(w, http.StatusOK, geminiSafetyResponse(ve))
			return
		}
		writeJSON(w, ve.Code, vertexErrorToGemini(ve))
		return
	}
	cleanGeminiFinishReason(resp)
	writeJSON(w, http.StatusOK, resp)
}

// handleGeminiStreamGenerate 纯原生实时流式直通，无聚合延迟、无下游中断隐患
func (g *GeminiHandler) handleGeminiStreamGenerate(w http.ResponseWriter, r *http.Request, model string) {
	actualModel, modelOK := resolveConfiguredModel(model, g.cfg)
	if !modelOK {
		geminiModelNotFound(w, model)
		return
	}
	body, ok := g.readGeminiBody(w, r)
	if !ok {
		return
	}
	if reqObj, ok2 := body["generateContentRequest"].(map[string]any); ok2 {
		body = reqObj
	}
	log.Printf("[Server] [GeminiStreamGenerate] 收到流式请求: 模型=%s, 真模型=%s", model, actualModel)

	sw := newSSEWriter(w, "text/event-stream")

	gotChunk := false
	hasFinish := false
	streamErrWritten := false
	g.vc.StreamChat(r.Context(), actualModel, body, func(ch vertex.StreamChunk) bool {
		if ch.Err != nil {
			if !sw.hasWritten() {
				ve := toVertexError(ch.Err)
				if isSafetyBlock(ve) {
					writeJSON(w, http.StatusOK, geminiSafetyResponse(ve))
				} else {
					writeJSON(w, ve.Code, vertexErrorToGemini(ve))
				}
			} else {
				if isSafetyBlock(ch.Err) {
					_ = sw.write(g.geminiSSE(geminiSafetyChunk(ch.Err)))
				} else {
					_ = sw.write(g.geminiSSE(map[string]any{"error": map[string]any{
						"code": ch.Err.Code, "message": vertex.FriendlyErrorMessage(ch.Err), "status": geminiStatusOf(ch.Err),
					}}))
				}
			}
			streamErrWritten = true
			return false
		}
		gotChunk = true
		if fr := cleanGeminiFinishReason(ch.Data); fr != "" {
			hasFinish = true
		}
		return sw.write(g.geminiSSE(ch.Data))
	})

	if streamErrWritten {
		return
	}

	if !gotChunk {
		ee := vertex.NewEmptyResponseError("Upstream returned empty response (no content)")
		if !sw.hasWritten() {
			writeJSON(w, ee.Code, vertexErrorToGemini(ee))
		} else {
			_ = sw.write(g.geminiSSE(map[string]any{
				"error": map[string]any{
					"code": 500, "message": "Upstream returned empty response (no content)", "status": "INTERNAL",
				},
			}))
		}
		return
	}
	if !hasFinish {
		_ = sw.write(g.geminiSSE(map[string]any{
			"candidates": []any{map[string]any{
				"content":      map[string]any{"parts": []any{}, "role": "model"},
				"finishReason": "STOP",
				"index":        0,
			}},
		}))
	}
}


func (g *GeminiHandler) handleModelInfo(w http.ResponseWriter, modelName string) {
	name := strings.TrimPrefix(modelName, "models/")
	if _, ok := resolveConfiguredModel(name, g.cfg); !ok {
		geminiModelNotFound(w, modelName)
		return
	}
	writeJSON(w, http.StatusOK, geminiModelInfo(name))
}

func (g *GeminiHandler) geminiSSE(obj map[string]any) string {
	data, err := jsonx.Marshal(obj)
	if err != nil {
		return "data: {}\n\n"
	}
	return "data: " + string(data) + "\n\n"
}

func (g *GeminiHandler) geminiError(w http.ResponseWriter, status int, msg, geminiStatus string) {
	writeJSON(w, status, map[string]any{"error": map[string]any{
		"code": status, "message": msg, "status": geminiStatus,
	}})
}

func cleanGeminiFinishReason(data map[string]any) string {
	cands, ok := data["candidates"].([]any)
	if !ok || len(cands) == 0 {
		return ""
	}
	var realFR string
	for _, candRaw := range cands {
		cand, ok := candRaw.(map[string]any)
		if !ok {
			continue
		}
		fr, _ := cand["finishReason"].(string)
		if fr == "FINISH_REASON_UNSPECIFIED" {
			delete(cand, "finishReason")
		} else if fr != "" {
			if realFR == "" {
				realFR = fr
			}
		}
	}
	return realFR
}

func vertexErrorToGemini(e *vertex.VertexError) map[string]any {
	msg := vertex.FriendlyErrorMessage(e)
	if e.Message != "" {
		msg += " | Raw: " + e.Message
	}
	if e.UpstreamResponse != "" {
		msg += " | Upstream: " + e.UpstreamResponse
	}
	return map[string]any{"error": map[string]any{
		"code": e.Code, "message": msg, "status": geminiStatusOf(e),
	}}
}

func geminiStatusOf(e *vertex.VertexError) string {
	if e != nil && e.Status != "" {
		return e.Status
	}
	return "INTERNAL"
}

func geminiSafetyResponse(e *vertex.VertexError) map[string]any {
	blockReason := e.Status
	if blockReason == "" {
		blockReason = "SAFETY"
	}
	return map[string]any{
		"candidates": []any{},
		"promptFeedback": map[string]any{
			"blockReason":        blockReason,
			"safetyRatings":      []any{},
			"blockReasonMessage": e.Message,
		},
	}
}

func geminiSafetyChunk(e *vertex.VertexError) map[string]any {
	blockReason := e.Status
	if blockReason == "" {
		blockReason = "SAFETY"
	}
	return map[string]any{
		"candidates": []any{map[string]any{
			"content":       map[string]any{"parts": []any{}, "role": "model"},
			"finishReason":  "SAFETY",
			"safetyRatings": []any{},
			"index":         0,
		}},
		"promptFeedback": map[string]any{
			"blockReason":        blockReason,
			"safetyRatings":      []any{},
			"blockReasonMessage": e.Message,
		},
	}
}
