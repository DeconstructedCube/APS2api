package config

type staticConfig struct {
	c AppConfig
}

func StaticProvider(c AppConfig) ConfigProvider { return staticConfig{c} }

func (s staticConfig) PortAPI() int                      { return s.c.PortAPI }
func (s staticConfig) MaxRetries() int                   { return s.c.MaxRetries }
func (s staticConfig) AdminPassword() string             { return s.c.AdminPassword }
func (s staticConfig) ProxyURL() string                  { return s.c.ProxyURL }
func (s staticConfig) DebugPprof() bool                  { return s.c.DebugPprof }
func (s staticConfig) DebugMode() bool                   { return s.c.DebugMode }
func (s staticConfig) MaxRequestMB() int                 { return s.c.MaxRequestMB }
func (s staticConfig) MaxSpillMB() int                   { return s.c.MaxSpillMB }
func (s staticConfig) RequestTimeout() int               { return s.c.RequestTimeout }
func (s staticConfig) RaceTimeout() int                  { return s.c.RaceTimeout }
func (s staticConfig) StreamIdleTimeoutSeconds() int     { return s.c.StreamIdleTimeoutSeconds }
func (s staticConfig) VertexAPIKey() string              { return s.c.VertexAPIKey }
func (s staticConfig) SafetySettings() map[string]string { return s.c.SafetySettings }
func (s staticConfig) ParallelPoolEnabled() bool         { return s.c.ParallelPoolEnabled }
func (s staticConfig) StickyNodePriority() bool          { return s.c.StickyNodePriority }
func (s staticConfig) ParallelPoolRetryEnabled() bool    { return s.c.ParallelPoolRetryEnabled }
func (s staticConfig) ParallelPoolSize() int             { return s.c.ParallelPoolSize }
func (s staticConfig) EntryProxyProbeEnabled() bool      { return s.c.EntryProxyProbeEnabled }
func (s staticConfig) EntryProxyProbeIntervalSeconds() int {
	return s.c.EntryProxyProbeIntervalSeconds
}
func (s staticConfig) EntryProxyProbeCooldownSeconds() int {
	return s.c.EntryProxyProbeCooldownSeconds
}
func (s staticConfig) EntryProxyProbeAutoDisableEnabled() bool {
	return s.c.EntryProxyProbeAutoDisableEnabled
}
func (s staticConfig) EntryProxyProbeAutoDisableFailures() int {
	return s.c.EntryProxyProbeAutoDisableFailures
}
func (s staticConfig) RecaptchaTryEntryOrDirect() bool          { return s.c.RecaptchaTryEntryOrDirect }
func (s staticConfig) ActiveNodeURI() string                    { return s.c.ActiveNodeURI }
func (s staticConfig) ParallelNodeTopK() int                    { return s.c.ParallelNodeTopK }
func (s staticConfig) AutoRefreshLogs() bool                    { return s.c.GetAutoRefreshLogs() }
func (s staticConfig) BaseModels() []string                     { return s.c.BaseModels() }
func (s staticConfig) ModelRegistry() []ModelEntry              { return s.c.ModelRegistry() }
func (s staticConfig) AliasMap() map[string]string              { return s.c.AliasMap() }
func (s staticConfig) ResolveModelName(s_ string) string        { return s.c.ResolveModelName(s_) }
func (s staticConfig) LookupModel(s_ string) (ModelEntry, bool) { return s.c.LookupModel(s_) }
func (s staticConfig) ConfigDir() string                        { return s.c.ConfigDir() }
func (s staticConfig) ConfigPath() string                       { return s.c.ConfigPath() }
