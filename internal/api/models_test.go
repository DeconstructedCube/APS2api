package api

import (
	"path/filepath"
	"testing"

	"AnonymousVertex/internal/config"
)

func TestResolveConfiguredModel(t *testing.T) {
	path := filepath.Join(t.TempDir(), "models.json")
	t.Setenv("VPROXY_MODELS", path)
	config.InvalidateModelsCache()
	if err := config.WriteModelRegistry([]config.ModelEntry{
		{ID: "enabled", Enabled: true},
		{ID: "disabled", Enabled: false},
	}, map[string]string{"alias": "enabled", "off-alias": "disabled"}); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(config.InvalidateModelsCache)

	provider := config.StaticProvider(config.DefaultConfig())
	cases := []struct {
		name, raw, want string
		ok              bool
	}{
		{"基础模型", "enabled", "enabled", true},
		{"别名", "alias", "enabled", true},
		{"禁用模型", "disabled", "disabled", false},
		{"禁用模型别名", "off-alias", "disabled", false},
		{"未知模型", "missing", "missing", false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, ok := resolveConfiguredModel(tc.raw, provider)
			if got != tc.want || ok != tc.ok {
				t.Fatalf("resolve(%q)=(%q,%v), want (%q,%v)", tc.raw, got, ok, tc.want, tc.ok)
			}
		})
	}
}
