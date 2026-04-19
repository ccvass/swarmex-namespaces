package namespaces

import "testing"

func TestParseConfig(t *testing.T) {
	tests := []struct {
		name   string
		labels map[string]string
		want   Config
	}{
		{
			"empty labels",
			map[string]string{},
			Config{},
		},
		{
			"namespace only",
			map[string]string{labelNamespace: "frontend"},
			Config{Namespace: "frontend", Isolated: false},
		},
		{
			"namespace isolated",
			map[string]string{labelNamespace: "backend", labelIsolated: "true"},
			Config{Namespace: "backend", Isolated: true},
		},
		{
			"isolated false",
			map[string]string{labelNamespace: "data", labelIsolated: "false"},
			Config{Namespace: "data", Isolated: false},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ParseConfig(tt.labels)
			if got != tt.want {
				t.Errorf("ParseConfig() = %+v, want %+v", got, tt.want)
			}
		})
	}
}
