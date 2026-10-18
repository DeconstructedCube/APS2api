package config

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"
)

const (
	defaultAnonAPIKey                     = "AIzaSyCI-zsRP85UVOi0DjtiCwWBwQ1djDy741g"
	maxTimeoutSeconds                     = 1800
	DefaultEntryProxyProbeIntervalSeconds = 300
	DefaultEntryProxyProbeCooldownSeconds = 60
	DefaultEntryProxyAutoDisableFailures  = 10
	DefaultRaceTimeoutSeconds             = 60
	DefaultStreamIdleTimeoutSeconds       = 20
	MinEntryProxyProbeIntervalSeconds     = 60
	MaxEntryProxyProbeSeconds             = 86400
	MaxEntryProxyAutoDisableFailures      = 100
)

type AppConfig struct { //nolint:govet
	PortAPI                   int               `json:"port_api"`
	MaxRetries                int               `json:"max_retries"`
	AdminPassword             string            `json:"admin_password"`
	ProxyURL                  string            `json:"proxy_url"`
	SafetySettings            map[string]string `json:"safety_settings"`
	VertexAPIKey              string            `json:"vertex_api_key"`
	MaxSpillMB                int               `json:"max_spill_mb"`
	MaxRequestMB              int               `json:"max_request_mb"`
	RequestTimeout            int               `json:"request_timeout"`
	RaceTimeout               int               `json:"race_timeout"`
	StreamIdleTimeoutSeconds  int               `json:"stream_idle_timeout_seconds"` // 流式包间空闲超时（秒），防呆连接假死

	// 并发池与节点锁定配置
	ActiveNodeURI                      string `json:"active_node_uri"`
	ParallelPoolEnabled                bool   `json:"parallel_pool_enabled"`
	StickyNodePriority                 bool   `json:"sticky_node_priority"`
	ParallelPoolRetryEnabled           bool   `json:"parallel_pool_retry_enabled"`
	ParallelPoolSize                   int    `json:"parallel_pool_size"`
	DebugPprof                         bool   `json:"debug_pprof"`
	ParallelNodeTopK                   int    `json:"parallel_node_top_k"`
	DebugMode                          bool   `json:"debug_mode"`
	EntryProxyProbeEnabled             bool   `json:"entry_proxy_probe_enabled"`
	EntryProxyProbeIntervalSeconds     int    `json:"entry_proxy_probe_interval_seconds"`
	EntryProxyProbeCooldownSeconds     int    `json:"entry_proxy_probe_cooldown_seconds"`
	EntryProxyProbeAutoDisableEnabled  bool   `json:"entry_proxy_probe_auto_disable_enabled"`
	EntryProxyProbeAutoDisableFailures int    `json:"entry_proxy_probe_auto_disable_failures"`
	RecaptchaTryEntryOrDirect          bool   `json:"recaptcha_try_entry_or_direct"`

}

func DefaultConfig() AppConfig {
	return AppConfig{ //nolint:exhaustruct
		PortAPI:                            2156,
		MaxRetries:                         1, // 默认为 1 次
		VertexAPIKey:                       defaultAnonAPIKey,
		MaxSpillMB:                         32,
		MaxRequestMB:                       128,
		RequestTimeout:                     180,
		RaceTimeout:                        DefaultRaceTimeoutSeconds,
		StreamIdleTimeoutSeconds:           DefaultStreamIdleTimeoutSeconds,
		ParallelPoolEnabled:                true,
		StickyNodePriority:                 false,
		ParallelPoolSize:                   15, // 默认为 15 并发
		ParallelNodeTopK:                   80,
		EntryProxyProbeEnabled:             false, // 周期拨测默认关闭，避免后台自动产生入口代理流量
		EntryProxyProbeIntervalSeconds:     DefaultEntryProxyProbeIntervalSeconds,
		EntryProxyProbeCooldownSeconds:     DefaultEntryProxyProbeCooldownSeconds,
		EntryProxyProbeAutoDisableEnabled:  false,
		EntryProxyProbeAutoDisableFailures: DefaultEntryProxyAutoDisableFailures,
		RecaptchaTryEntryOrDirect:          true,
	}
}

var (
	//nolint:gochecknoglobals // Global configuration cache
	mu sync.Mutex
	//nolint:gochecknoglobals // writeMu serializes config.json read-modify-write operations.
	writeMu sync.Mutex
	//nolint:gochecknoglobals // Global configuration cache
	cached *AppConfig
	//nolint:gochecknoglobals // Global configuration cache
	cacheTime time.Time
	//nolint:gochecknoglobals // Tracks successful loads so unchanged TTL refreshes stay quiet.
	lastLoadedConfigHash [sha256.Size]byte
	//nolint:gochecknoglobals // Tracks whether lastLoadedConfigHash has been initialized.
	hasLoadedConfigHash bool
)

const cacheTTL = 60 * time.Second

func configPath() string {
	if p := os.Getenv("VPROXY_CONFIG"); p != "" {
		return p
	}
	if exe, err := os.Executable(); err == nil {
		p := filepath.Join(filepath.Dir(exe), "config", "config.json")
		if _, errStat := os.Stat(p); errStat == nil { //nolint:govet
			return p
		}
	}
	return filepath.Join("config", "config.json")
}

func ConfigPath() string { return configPath() }

func ConfigDir() string { return filepath.Dir(configPath()) }

func WriteSettings(updates map[string]any) error {
	if err := writeSettings(updates); err != nil {
		return err
	}
	InvalidateCache()
	return nil
}

func writeSettings(updates map[string]any) error {
	writeMu.Lock()
	defer writeMu.Unlock()

	path := configPath()
	raw := map[string]any{}
	if data, err := os.ReadFile(path); err == nil {
		_ = json.Unmarshal(data, &raw)
	}
	for k, v := range updates {
		raw[k] = v
	}

	// 拦截并在面板保存配置时限制并发上限为 20
	if val, ok := raw["parallel_pool_size"].(float64); ok && val > 20 {
		log.Printf("[Config] 面板设置并发数过高 (%v)，已强制保存为上限 20", val)
		raw["parallel_pool_size"] = 20
	} else if val, ok2 := raw["parallel_pool_size"].(int); ok2 && val > 20 { //nolint:govet
		log.Printf("[Config] 面板设置并发数过高 (%v)，已强制保存为上限 20", val)
		raw["parallel_pool_size"] = 20
	}

	if err := writeJSONFile(path, raw); err != nil {
		return err
	}
	return nil
}

func writeJSONFile(path string, v any) error {
	if dir := filepath.Dir(path); dir != "" {
		_ = os.MkdirAll(dir, 0o755)
	}
	data, _ := json.MarshalIndent(v, "", "  ")
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return fmt.Errorf("write temp config %s: %w", tmp, err)
	}
	return os.Rename(tmp, path) //nolint:wrapcheck
}

func Load() AppConfig {
	mu.Lock()
	defer mu.Unlock()
	if cached != nil && time.Since(cacheTime) < cacheTTL {
		return *cached
	}
	cfg := DefaultConfig()
	initializeConfigFromExample()
	if data, err := os.ReadFile(configPath()); err == nil {
		if errUnm := json.Unmarshal(data, &cfg); errUnm != nil { //nolint:govet
			log.Printf("[Config] 解析 config.json 失败: %v", err)
		} else {
			if updates := normalizeLoadedConfig(&cfg); len(updates) > 0 {
				// 在返回配置前完成回写，避免调用方读到已修正值但文件仍保留非法值。
				if errSave := writeSettings(updates); errSave != nil {
					log.Printf("[Config] 自动回写规范化配置失败: %v", errSave)
				}
			}
			if shouldLogSuccessfulLoad(cfg) {
				log.Printf("[Config] 成功加载配置文件 config.json")
			}
		}
	} else if !os.IsNotExist(err) {
		log.Printf("[Config] 读取 config.json 失败: %v", err)
	}
	cached = &cfg
	cacheTime = time.Now()
	return cfg
}

// shouldLogSuccessfulLoad suppresses the periodic success message when the
// cache TTL expires but the effective configuration has not changed.
// Load holds mu while calling this function.
func shouldLogSuccessfulLoad(cfg AppConfig) bool {
	data, err := json.Marshal(cfg)
	if err != nil {
		return true
	}
	hash := sha256.Sum256(data)
	if hasLoadedConfigHash && hash == lastLoadedConfigHash {
		return false
	}
	lastLoadedConfigHash = hash
	hasLoadedConfigHash = true
	return true
}

func initializeConfigFromExample() {
	path := configPath()
	if _, err := os.Stat(path); err == nil || !os.IsNotExist(err) {
		return
	}
	examplePath := filepath.Join(filepath.Dir(path), "config.example.json")
	data, err := os.ReadFile(examplePath)
	if err != nil {
		if !os.IsNotExist(err) {
			log.Printf("[Config] 读取初始配置模板失败: %v", err)
		}
		return
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		log.Printf("[Config] 创建配置目录失败: %v", err)
		return
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		log.Printf("[Config] 从 config.example.json 创建初始配置失败: %v", err)
		return
	}
	log.Printf("[Config] 已从 config.example.json 创建初始配置")
}

func InvalidateCache() {
	mu.Lock()
	defer mu.Unlock()
	cached = nil
}

func (c AppConfig) ConfigDir() string  { return ConfigDir() }
func (c AppConfig) ConfigPath() string { return ConfigPath() }

func (c *AppConfig) WriteSettings(updates map[string]any) error { return WriteSettings(updates) }
func (c *AppConfig) WriteModels(models []string, aliasMap map[string]string) error {
	return WriteModels(models, aliasMap)
}

func normalizeLoadedConfig(cfg *AppConfig) map[string]any {
	var needsSave bool
	if cfg.RequestTimeout <= 0 {
		cfg.RequestTimeout = 180
		needsSave = true
	} else if cfg.RequestTimeout > maxTimeoutSeconds {
		log.Printf("[Config] 警告: 请求超时配置过高 (%d)，已限制为上限 %d", cfg.RequestTimeout, maxTimeoutSeconds)
		cfg.RequestTimeout = maxTimeoutSeconds
		needsSave = true
	}
	if cfg.RaceTimeout < 0 {
		cfg.RaceTimeout = 0
		needsSave = true
	} else if cfg.RaceTimeout > maxTimeoutSeconds {
		log.Printf("[Config] 警告: 竞速超时配置过高 (%d)，已限制为上限 %d", cfg.RaceTimeout, maxTimeoutSeconds)
		cfg.RaceTimeout = maxTimeoutSeconds
		needsSave = true
	}
	if cfg.StreamIdleTimeoutSeconds <= 0 {
		cfg.StreamIdleTimeoutSeconds = DefaultStreamIdleTimeoutSeconds
		needsSave = true
	}
	if cfg.EntryProxyProbeIntervalSeconds <= 0 {
		cfg.EntryProxyProbeIntervalSeconds = DefaultEntryProxyProbeIntervalSeconds
		needsSave = true
	} else if cfg.EntryProxyProbeIntervalSeconds < MinEntryProxyProbeIntervalSeconds {
		cfg.EntryProxyProbeIntervalSeconds = MinEntryProxyProbeIntervalSeconds
		needsSave = true
	} else if cfg.EntryProxyProbeIntervalSeconds > MaxEntryProxyProbeSeconds {
		cfg.EntryProxyProbeIntervalSeconds = MaxEntryProxyProbeSeconds
		needsSave = true
	}
	if cfg.EntryProxyProbeCooldownSeconds < 0 {
		cfg.EntryProxyProbeCooldownSeconds = 0
		needsSave = true
	} else if cfg.EntryProxyProbeCooldownSeconds > MaxEntryProxyProbeSeconds {
		cfg.EntryProxyProbeCooldownSeconds = MaxEntryProxyProbeSeconds
		needsSave = true
	}
	if cfg.EntryProxyProbeAutoDisableFailures <= 0 {
		cfg.EntryProxyProbeAutoDisableFailures = DefaultEntryProxyAutoDisableFailures
		needsSave = true
	} else if cfg.EntryProxyProbeAutoDisableFailures > MaxEntryProxyAutoDisableFailures {
		cfg.EntryProxyProbeAutoDisableFailures = MaxEntryProxyAutoDisableFailures
		needsSave = true
	}
	if cfg.ParallelPoolSize > 20 {
		log.Printf("[Config] 警告: 并发数配置过高 (%d)，已限制为上限 20", cfg.ParallelPoolSize)
		cfg.ParallelPoolSize = 20
		needsSave = true
	}
	if !needsSave {
		return nil
	}
	return map[string]any{
		"request_timeout":                         cfg.RequestTimeout,
		"race_timeout":                            cfg.RaceTimeout,
		"stream_idle_timeout_seconds":             cfg.StreamIdleTimeoutSeconds,
		"parallel_pool_size":                      cfg.ParallelPoolSize,
		"entry_proxy_probe_interval_seconds":      cfg.EntryProxyProbeIntervalSeconds,
		"entry_proxy_probe_cooldown_seconds":      cfg.EntryProxyProbeCooldownSeconds,
		"entry_proxy_probe_auto_disable_failures": cfg.EntryProxyProbeAutoDisableFailures,
		"recaptcha_try_entry_or_direct":           cfg.RecaptchaTryEntryOrDirect,
	}
}

