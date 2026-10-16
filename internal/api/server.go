// Package api 暴露 Gemini 原生 HTTP 端点与管理控制台。
package api

import (
	"context"
	"log"
	"net/http"
	"path/filepath"
	"sync"
	"time"

	"AnonymousVertex/internal/buildinfo"
	"AnonymousVertex/internal/config"
	"AnonymousVertex/internal/nodes"
	"AnonymousVertex/internal/subscriptions"
	"AnonymousVertex/internal/vertex"
)

type Server struct {
	gemini            *GeminiHandler
	admin             *AdminHandler
	subscriptions     *subscriptions.Service
	mw                *middleware
	backgroundMu      sync.Mutex
	backgroundCancels map[string]context.CancelFunc
	backgroundWG      sync.WaitGroup
}

func NewServer(vc *vertex.VertexAIClient, keys *APIKeyManager, cfg config.ConfigProvider) *Server {
	h := handler{vc: vc, keys: keys, cfg: cfg}
	adminHandler := &AdminHandler{handler: h}
	subscriptionService := subscriptions.New(func(ctx context.Context, rawURL, userAgent string) ([]nodes.Node, error) {
		text, err := adminHandler.fetchSubWithFallback(ctx, rawURL, userAgent)
		if err != nil {
			return nil, err
		}
		return parseImportedNodes(text), nil
	})
	adminHandler.subscriptionService = subscriptionService
	srv := &Server{
		gemini:            &GeminiHandler{handler: h},
		admin:             adminHandler,
		subscriptions:     subscriptionService,
		mw:                &middleware{cfg: cfg, keys: keys},
		backgroundCancels: make(map[string]context.CancelFunc),
	}

	if err := subscriptionService.Start(context.Background()); err != nil {
		log.Printf("[Subscriptions] 启动自动更新服务失败: %v", err)
	}
	return srv
}

func (s *Server) Close() {
	s.backgroundMu.Lock()
	for _, cancel := range s.backgroundCancels {
		cancel()
	}
	s.backgroundMu.Unlock()
	s.backgroundWG.Wait()

	if s.subscriptions != nil {
		s.subscriptions.Stop()
	}
}

func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/", s.handleRoot)
	mux.HandleFunc("/health", s.handleHealth)
	mux.HandleFunc("/api/meta/build", s.handleBuildMeta)

	// 模型列表
	mux.HandleFunc("/v1beta/models", s.handleModelsGemini)

	// Gemini 端点子树
	mux.HandleFunc("/v1beta/models/", s.gemini.handleModelsSubtree)

	// 管理控制台与静态资源
	mux.HandleFunc("/favicon.ico", s.handleFavicon)
	mux.HandleFunc("/admin", s.admin.handleAdminPage)
	mux.HandleFunc("/admin/", s.admin.handleAdminPage)
	mux.HandleFunc("/api/admin/", s.admin.handleAdminAPI)
	mux.HandleFunc("/assets/", s.handleAssets)

	if s.mw.cfg.DebugPprof() {
		mux.HandleFunc("/debug/pprof/", pprofIndex)
		mux.HandleFunc("/debug/pprof/cmdline", pprintCmdline)
		mux.HandleFunc("/debug/pprof/profile", pprofProfile)
		mux.HandleFunc("/debug/pprof/symbol", pprofSymbol)
		mux.HandleFunc("/debug/pprof/trace", pprofTrace)
		mux.HandleFunc("/debug/pprof/goroutine", pprofGoroutine)
		mux.HandleFunc("/debug/pprof/heap", pprofHeap)
		mux.HandleFunc("/debug/pprof/threadcreate", pprofThreadcreate)
		mux.HandleFunc("/debug/pprof/block", pprofBlock)
		mux.HandleFunc("/debug/pprof/mutex", pprofMutex)
	}

	return s.mw.withRecover(s.mw.withCORS(s.mw.withMetrics(s.mw.withAPIKey(s.mw.withBodyLimit(mux)))))
}

func (s *Server) handleFavicon(w http.ResponseWriter, _ *http.Request) {
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) handleAssets(w http.ResponseWriter, r *http.Request) {
	assetsDir := filepath.Join(filepath.Dir(s.mw.cfg.ConfigDir()), "assets")
	fs := http.StripPrefix("/assets/", http.FileServer(http.Dir(assetsDir)))
	fs.ServeHTTP(w, r)
}

func (s *Server) handleRoot(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		writeJSON(w, http.StatusNotFound, map[string]any{"error": map[string]any{"message": "not found", "status": "NOT_FOUND"}})
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"message": "AnonymousVertex Gemini Proxy", "version": buildinfo.Current().Version})
}

func (s *Server) handleHealth(w http.ResponseWriter, _ *http.Request) {
	log.Printf("[Server] [Health] 收到健康检查请求")
	writeJSON(w, http.StatusOK, map[string]any{
		"status":          "healthy",
		"timestamp":       time.Now().Unix(),
		"api_keys_loaded": s.mw.keys.Count(),
	})
}

func (s *Server) handleBuildMeta(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")
	if r.Method != http.MethodGet {
		apiError(w, http.StatusMethodNotAllowed, "method not allowed", "METHOD_NOT_ALLOWED")
		return
	}
	writeJSON(w, http.StatusOK, buildinfo.Current())
}

func (s *Server) handleModelsGemini(w http.ResponseWriter, _ *http.Request) {
	models := s.mw.cfg.BaseModels()
	data := make([]any, 0, len(models))
	for _, m := range models {
		data = append(data, geminiModelInfo(m))
	}
	writeJSON(w, http.StatusOK, map[string]any{"models": data})
}
