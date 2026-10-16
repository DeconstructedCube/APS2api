package api

import (
	"net/http"

	"AnonymousVertex/internal/config"
)

//nolint:gochecknoglobals // Constant-like map of allowed settings
var adminAllowedSettings = map[string]bool{
	"max_retries": true, "max_spill_mb": true,
	"max_request_mb": true, "max_n": true,
	"request_timeout":               true,
	"race_timeout":                  true,
	"stream_idle_timeout_seconds":   true,
	"recaptcha_try_entry_or_direct": true,
	"parallel_pool_enabled":         true, "parallel_pool_size": true,
	"parallel_pool_delay_dynamic":             true,
	"entry_proxy_probe_enabled":               true,
	"entry_proxy_probe_interval_seconds":      true,
	"entry_proxy_probe_cooldown_seconds":      true,
	"entry_proxy_probe_auto_disable_enabled":  true,
	"entry_proxy_probe_auto_disable_failures": true,
	"parallel_pool_delay_ms":                  true,
	"active_node_uri":                         true,
	"sticky_node_priority":                    true,
	"parallel_pool_retry_enabled":             true,
	"debug_mode":                              true,
}

func (adm *AdminHandler) adminGetSettings(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{"settings": map[string]any{
		"max_retries":                   adm.cfg.MaxRetries(),
		"max_spill_mb":                  adm.cfg.MaxSpillMB(),
		"max_request_mb":                adm.cfg.MaxRequestMB(),
		"max_n":                         adm.cfg.MaxN(),
		"request_timeout":               adm.cfg.RequestTimeout(),
		"race_timeout":                  adm.cfg.RaceTimeout(),
		"stream_idle_timeout_seconds":   adm.cfg.StreamIdleTimeoutSeconds(),
		"recaptcha_try_entry_or_direct": adm.cfg.RecaptchaTryEntryOrDirect(),
		"proxy_url":                     adm.cfg.ProxyURL(),
		"parallel_pool_enabled":         adm.cfg.ParallelPoolEnabled(),
		"parallel_pool_size":            adm.cfg.ParallelPoolSize(),
		"active_node_uri":               adm.cfg.ActiveNodeURI(),
		"proxy_url_candidates":          config.ListProxyCandidates(),
		"parallel_pool_delay_dynamic":   adm.cfg.ParallelPoolDelayDynamic(),
		"parallel_pool_delay_ms":        adm.cfg.ParallelPoolDelayMs(),
		"entry_proxy_probe_enabled":               adm.cfg.EntryProxyProbeEnabled(),
		"entry_proxy_probe_interval_seconds":      adm.cfg.EntryProxyProbeIntervalSeconds(),
		"entry_proxy_probe_cooldown_seconds":      adm.cfg.EntryProxyProbeCooldownSeconds(),
		"entry_proxy_probe_auto_disable_enabled":  adm.cfg.EntryProxyProbeAutoDisableEnabled(),
		"entry_proxy_probe_auto_disable_failures": adm.cfg.EntryProxyProbeAutoDisableFailures(),
		"sticky_node_priority":                    adm.cfg.StickyNodePriority(),
		"parallel_pool_retry_enabled":             adm.cfg.ParallelPoolRetryEnabled(),
		"debug_mode":                              adm.cfg.DebugMode(),
		"auto_refresh_logs":                       true,
	}})
}

func (adm *AdminHandler) adminPutSettings(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Settings map[string]any `json:"settings"`
	}
	if !adm.decodeAdminBody(w, r, &body) {
		return
	}
	updates := map[string]any{}

	// 面板依赖校验：禁用并发池时强制禁用粘性池
	if ppEnabled, ok := body.Settings["parallel_pool_enabled"].(bool); ok && !ppEnabled {
		body.Settings["sticky_node_priority"] = false
	}

	for k, v := range body.Settings {
		if !adminAllowedSettings[k] {
			continue
		}
		switch k {
		case "max_retries", "max_spill_mb", "max_request_mb", "max_n", "parallel_pool_size", "parallel_pool_delay_ms", "request_timeout", "race_timeout", "stream_idle_timeout_seconds",
			"entry_proxy_probe_interval_seconds", "entry_proxy_probe_cooldown_seconds", "entry_proxy_probe_auto_disable_failures":
			if f, ok := v.(float64); ok {
				val := int(f)
				switch k {
				case "request_timeout":
					if val <= 0 {
						val = 180
					} else if val > 1800 {
						val = 1800
					}
				case "race_timeout":
					if val < 0 {
						val = 0
					} else if val > 1800 {
						val = 1800
					}
				case "stream_idle_timeout_seconds":
					if val <= 0 {
						val = config.DefaultStreamIdleTimeoutSeconds
					} else if val > 1800 {
						val = 1800
					}
				case "entry_proxy_probe_interval_seconds":
					if val <= 0 {
						val = config.DefaultEntryProxyProbeIntervalSeconds
					} else if val < config.MinEntryProxyProbeIntervalSeconds {
						val = config.MinEntryProxyProbeIntervalSeconds
					} else if val > config.MaxEntryProxyProbeSeconds {
						val = config.MaxEntryProxyProbeSeconds
					}
				case "entry_proxy_probe_cooldown_seconds":
					if val < 0 {
						val = 0
					} else if val > config.MaxEntryProxyProbeSeconds {
						val = config.MaxEntryProxyProbeSeconds
					}
				case "entry_proxy_probe_auto_disable_failures":
					if val <= 0 {
						val = config.DefaultEntryProxyAutoDisableFailures
					} else if val > config.MaxEntryProxyAutoDisableFailures {
						val = config.MaxEntryProxyAutoDisableFailures
					}
				}
				updates[k] = val
				continue
			}
		}
		updates[k] = v
	}
	if err := config.WriteSettings(updates); err != nil {
		writeJSON(w, http.StatusInternalServerError, adminErr("写入配置失败 (failed to write config)"))
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}

func (adm *AdminHandler) adminGetStats(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, metricsBody())
}

func (adm *AdminHandler) adminResetStats(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{"ok": true})
}
