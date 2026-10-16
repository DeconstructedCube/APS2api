package api

import (
	"context"
	cryptorand "crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"AnonymousVertex/internal/config"
	"AnonymousVertex/internal/jsonx"
	"AnonymousVertex/internal/strutil"
	"AnonymousVertex/internal/vertex"
)

type handler struct {
	vc   *vertex.VertexAIClient
	keys *APIKeyManager
	cfg  config.ConfigProvider
}

func (h *handler) decodeAdminBody(w http.ResponseWriter, r *http.Request, dst any) bool {
	if r.Body == nil {
		writeJSON(w, http.StatusBadRequest, adminErr("请求体为空 (empty body)"))
		return false
	}
	if err := json.NewDecoder(r.Body).Decode(dst); err != nil {
		writeJSON(w, http.StatusBadRequest, adminErr("请求格式错误 (invalid JSON)"))
		return false
	}
	return true
}

func (h *handler) adminUnauthorized(w http.ResponseWriter) {
	writeJSON(w, http.StatusUnauthorized, adminErr("未登录或会话已过期 (unauthorized)"))
}

func (h *handler) adminMethodNotAllowed(w http.ResponseWriter) {
	writeJSON(w, http.StatusMethodNotAllowed, adminErr("方法不允许 (method not allowed)"))
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	if status < 100 || status > 599 {
		status = http.StatusInternalServerError
	}
	data, err := jsonx.Marshal(body)
	if err != nil {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(`{"error":{"message":"序列化失败 (internal error)","status":"INTERNAL"}}`))
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_, _ = w.Write(data)
}

func apiError(w http.ResponseWriter, status int, msg, statusText string) {
	if statusText == "" {
		switch status {
		case http.StatusBadRequest:
			statusText = "INVALID_ARGUMENT"
		case http.StatusUnauthorized:
			statusText = "UNAUTHENTICATED"
		case http.StatusForbidden:
			statusText = "PERMISSION_DENIED"
		case http.StatusNotFound:
			statusText = "NOT_FOUND"
		default:
			statusText = "INTERNAL"
		}
	}
	writeJSON(w, status, map[string]any{"error": map[string]any{
		"code":    status,
		"message": msg,
		"status":  statusText,
	}})
}

type sseWriter struct {
	w           http.ResponseWriter
	contentType string
	written     bool
	flush       func()
}

func (s *sseWriter) hasWritten() bool {
	return s.written
}

func (s *sseWriter) write(data string) bool {
	if !s.written {
		s.w.Header().Set("Content-Type", s.contentType)
		s.w.Header().Set("Cache-Control", "no-cache")
		s.w.Header().Set("Connection", "keep-alive")
		s.w.WriteHeader(http.StatusOK)
		s.written = true
	}
	_, err := s.w.Write([]byte(data))
	if err != nil {
		return false
	}
	if s.flush != nil {
		s.flush()
	}
	return true
}

func newSSEWriter(w http.ResponseWriter, contentType string) *sseWriter {
	flusher, _ := w.(http.Flusher)
	sw := &sseWriter{
		w:           w,
		contentType: contentType,
	}
	if flusher != nil {
		sw.flush = flusher.Flush
	}
	return sw
}

func reqID24() string {
	return strutil.ReqID()
}

func toVertexError(err error) *vertex.VertexError {
	var ve *vertex.VertexError
	if errors.As(err, &ve) {
		return ve
	}
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return vertex.NewContextError(err)
	}
	return vertex.NewInternalError(err.Error(), err)
}

func isSafetyBlock(e *vertex.VertexError) bool {
	if e == nil {
		return false
	}
	if e.Kind == "safety" {
		return true
	}
	msg := strings.ToLower(e.Message)
	status := strings.ToLower(e.Status)
	for _, k := range []string{"safety", "block_reason", "content_filter", "finish_reason_safety"} {
		if strings.Contains(msg, k) || strings.Contains(status, k) {
			return true
		}
	}
	return false
}

func adminErr(msg string) map[string]any {
	return map[string]any{"error": map[string]any{"message": msg}}
}

func maskKey(key string) string {
	if len(key) <= 4 {
		return "sk-····"
	}
	return "sk-····" + key[len(key)-4:]
}

func generateAPIKey() string {
	b := make([]byte, 24)
	_, _ = cryptorand.Read(b)
	return "sk-" + hex.EncodeToString(b)
}
