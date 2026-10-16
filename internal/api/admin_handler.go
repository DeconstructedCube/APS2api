package api

import (
	"io/fs"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"AnonymousVertex/internal/admin"
	"AnonymousVertex/internal/subscriptions"
)

type AdminHandler struct {
	handler
	subscriptionService *subscriptions.Service
}

func (adm *AdminHandler) handleAdminAPI(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")
	path := strings.TrimPrefix(r.URL.Path, "/api/admin")
	log.Printf("[Server] [AdminAPI] 收到请求: %s %s", r.Method, path)

	if path == "/login" {
		adm.adminLogin(w, r)
		return
	}
	if path == "/check-auth" {
		adm.adminCheckAuth(w, r)
		return
	}

	if strings.HasPrefix(path, "/keys/") {
		if !requireAdmin(r) {
			adm.adminUnauthorized(w)
			return
		}
		adm.adminDeleteKey(w, r, strings.TrimPrefix(path, "/keys/"))
		return
	}

	if !requireAdmin(r) {
		adm.adminUnauthorized(w)
		return
	}

	switch path {
	case "/nodes":
		switch r.Method {
		case http.MethodGet:
			adm.adminGetNodes(w, r)
		case http.MethodDelete:
			adm.adminDeleteNode(w, r)
		}
		return
	case "/nodes/test":
		adm.adminTestNode(w, r)
		return
	case "/nodes/enable":
		adm.adminEnableNode(w, r)
		return
	case "/nodes/test-all":
		adm.adminTestAll(w, r)
		return
	case "/nodes/test-progress":
		if r.Method == http.MethodGet {
			adm.adminGetTestProgress(w, r)
		}
		return
	case "/nodes/test-pause":
		adm.adminTestPause(w, r)
		return
	case "/nodes/test-resume":
		adm.adminTestResume(w, r)
		return
	case "/nodes/test-terminate":
		adm.adminTestTerminate(w, r)
		return
	case "/nodes/deduplicate":
		if r.Method == http.MethodPost {
			adm.adminDedupNodes(w, r)
		} else {
			adm.adminMethodNotAllowed(w)
		}
		return
	case "/nodes/deduplicate/preview":
		if r.Method == http.MethodGet {
			adm.adminPreviewDedupNodes(w, r)
		} else {
			adm.adminMethodNotAllowed(w)
		}
		return
	case "/nodes/disabled":
		adm.adminDeleteDisabledNodes(w, r)
		return
	case "/nodes/import":
		adm.adminImportNodes(w, r)
		return
	case "/nodes/import-json":
		adm.adminImportNodesJson(w, r)
		return
	case "/subscriptions/fetch":
		if r.Method == http.MethodPost {
			adm.adminFetchSub(w, r)
		} else {
			adm.adminMethodNotAllowed(w)
		}
		return
	case "/subscriptions/list":
		if r.Method == http.MethodGet {
			adm.adminListSubscriptions(w, r)
		} else {
			adm.adminMethodNotAllowed(w)
		}
		return
	case "/subscriptions/save":
		if r.Method == http.MethodPost {
			adm.adminSaveSubscription(w, r)
		} else {
			adm.adminMethodNotAllowed(w)
		}
		return
	case "/subscriptions/delete":
		if r.Method == http.MethodPost {
			adm.adminDeleteSubscription(w, r)
		} else {
			adm.adminMethodNotAllowed(w)
		}
		return
	case "/subscriptions/update":
		if r.Method == http.MethodPost {
			adm.adminUpdateSubscriptions(w, r)
		} else {
			adm.adminMethodNotAllowed(w)
		}
		return
	case "/subscriptions/custom_ua/save":
		if r.Method == http.MethodPost {
			adm.adminSaveCustomUA(w, r)
		} else {
			adm.adminMethodNotAllowed(w)
		}
		return
	case "/subscriptions/custom_ua/delete":
		if r.Method == http.MethodPost {
			adm.adminDeleteCustomUA(w, r)
		} else {
			adm.adminMethodNotAllowed(w)
		}
		return
	case "/use-node":
		adm.adminUseNode(w, r)
		return
	case "/nodes/batch-disable":
		adm.adminBatchDisableNodes(w, r)
		return
	case "/nodes/batch-enable":
		adm.adminBatchEnableNodes(w, r)
		return
	case "/nodes/batch-delete":
		adm.adminBatchDeleteNodes(w, r)
		return
	case "/nodes/sort":
		adm.adminSortNodesByLatency(w, r)
		return
	case "/proxy-nodes/import":
		if r.Method == http.MethodPost {
			adm.adminImportProxyNode(w, r)
		} else {
			adm.adminMethodNotAllowed(w)
		}
		return
	case "/proxy-nodes/import-batch":
		if r.Method == http.MethodPost {
			adm.adminImportProxyNodesBatch(w, r)
		} else {
			adm.adminMethodNotAllowed(w)
		}
		return
	case "/proxy-nodes/enable-batch":
		if r.Method == http.MethodPost {
			adm.adminSetProxyNodesEnabled(w, r, true)
		} else {
			adm.adminMethodNotAllowed(w)
		}
		return
	case "/proxy-nodes/disable-batch":
		if r.Method == http.MethodPost {
			adm.adminSetProxyNodesEnabled(w, r, false)
		} else {
			adm.adminMethodNotAllowed(w)
		}
		return
	case "/proxy-nodes/delete-batch":
		if r.Method == http.MethodPost {
			adm.adminDeleteProxyNodesBatch(w, r)
		} else {
			adm.adminMethodNotAllowed(w)
		}
		return
	case "/proxy-nodes/disabled":
		if r.Method == http.MethodDelete {
			adm.adminDeleteDisabledProxyNodes(w, r)
		} else {
			adm.adminMethodNotAllowed(w)
		}
		return
	case "/proxy-nodes/enable":
		if r.Method == http.MethodPost {
			adm.adminEnableProxyNode(w, r)
		} else {
			adm.adminMethodNotAllowed(w)
		}
		return
	case "/proxy-nodes/disable":
		if r.Method == http.MethodPost {
			adm.adminDisableProxyNode(w, r)
		} else {
			adm.adminMethodNotAllowed(w)
		}
		return
	case "/proxy-nodes/test":
		if r.Method == http.MethodPost {
			adm.adminTestProxyNode(w, r)
		} else {
			adm.adminMethodNotAllowed(w)
		}
		return
	case "/proxy-nodes/test-batch":
		if r.Method == http.MethodPost {
			adm.adminBatchTestProxyNodes(w, r)
		} else {
			adm.adminMethodNotAllowed(w)
		}
		return
	case "/proxy-nodes/test-progress":
		if r.Method == http.MethodGet {
			adm.adminGetProxyTestProgress(w, r)
		} else {
			adm.adminMethodNotAllowed(w)
		}
		return
	case "/proxy-nodes":
		if r.Method == http.MethodGet {
			adm.adminListProxyNodes(w, r)
		} else if r.Method == http.MethodDelete {
			adm.adminDeleteProxyNode(w, r)
		} else {
			adm.adminMethodNotAllowed(w)
		}
		return
	}

	switch path {
	case "/logout":
		adm.adminLogout(w, r)
	case "/password":
		if r.Method == http.MethodPost {
			adm.adminChangePassword(w, r)
		} else {
			adm.adminMethodNotAllowed(w)
		}
	case "/settings":
		switch r.Method {
		case http.MethodGet:
			adm.adminGetSettings(w, r)
		case http.MethodPut:
			adm.adminPutSettings(w, r)
		default:
			adm.adminMethodNotAllowed(w)
		}
	case "/stats":
		adm.adminGetStats(w, r)
	case "/stats/reset":
		adm.adminResetStats(w, r)
	case "/log":
		if r.Method != http.MethodGet {
			adm.adminMethodNotAllowed(w)
			return
		}
		adm.adminGetLog(w, r)
	case "/keys":
		switch r.Method {
		case http.MethodGet:
			adm.adminGetKeys(w, r)
		case http.MethodPost:
			adm.adminAddKey(w, r)
		default:
			adm.adminMethodNotAllowed(w)
		}
	case "/models":
		switch r.Method {
		case http.MethodGet:
			adm.adminGetModels(w, r)
		case http.MethodPut:
			adm.adminPutModels(w, r)
		default:
			adm.adminMethodNotAllowed(w)
		}
	default:
		writeJSON(w, http.StatusNotFound, adminErr("未知接口 (not found)"))
	}
}

func (adm *AdminHandler) handleAdminPage(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/admin" {
		http.Redirect(w, r, "/admin/", http.StatusFound)
		return
	}
	name := strings.TrimPrefix(r.URL.Path, "/admin/")
	if name == "" {
		name = "index.html"
	}
	data, err := fs.ReadFile(admin.Assets, "assets/"+name)
	if err != nil && (name == "admin.html" || name == "index.html") {
		alt := "index.html"
		if name == "index.html" {
			alt = "admin.html"
		}
		data, err = fs.ReadFile(admin.Assets, "assets/"+alt)
	}
	if err != nil {
		apiError(w, http.StatusNotFound, "not found", "NOT_FOUND")
		return
	}
	w.Header().Set("Content-Type", contentTypeFor(name))
	w.Header().Set("Cache-Control", "no-store, no-cache, must-revalidate")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")
	_, _ = w.Write(data)
}

func contentTypeFor(name string) string {
	switch {
	case strings.HasSuffix(name, ".html"):
		return "text/html; charset=utf-8"
	case strings.HasSuffix(name, ".jpg"), strings.HasSuffix(name, ".jpeg"):
		return "image/jpeg"
	case strings.HasSuffix(name, ".png"):
		return "image/png"
	case strings.HasSuffix(name, ".css"):
		return "text/css; charset=utf-8"
	case strings.HasSuffix(name, ".js"):
		return "text/javascript; charset=utf-8"
	default:
		return "application/octet-stream"
	}
}

func (adm *AdminHandler) adminGetLog(w http.ResponseWriter, r *http.Request) {
	logPath := filepath.Join(filepath.Dir(adm.cfg.ConfigDir()), "logs", "logs_latest.log")

	if _, err := os.Stat(logPath); os.IsNotExist(err) {
		logPath = "logs/logs_latest.log"
	}

	data, err := os.ReadFile(logPath)
	if err != nil {
		if os.IsNotExist(err) {
			writeJSON(w, http.StatusOK, map[string]any{"ok": true, "content": ""})
			return
		}
		writeJSON(w, http.StatusInternalServerError, adminErr("无法读取日志文件 (read error)"))
		return
	}

	lines := strings.Split(string(data), "\n")
	var validLines []string
	for _, l := range lines {
		if strings.TrimSpace(l) != "" {
			validLines = append(validLines, l)
		}
	}
	if len(validLines) > 200 {
		validLines = validLines[len(validLines)-200:]
	}

	writeJSON(w, http.StatusOK, map[string]any{"ok": true, "content": strings.Join(validLines, "\n")})
}
