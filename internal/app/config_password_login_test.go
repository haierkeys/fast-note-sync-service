package app

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestLoadConfigPasswordLogin(t *testing.T) {
	for _, tc := range []struct {
		name, yaml string
		want       bool
	}{
		{"absent", "{}", true},
		{"empty", "user:\n  password-login-is-enable:\n", true},
		{"enabled", "user:\n  password-login-is-enable: true\n", true},
		{"disabled", "user:\n  password-login-is-enable: false\n", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			cfg, path, err := LoadConfig(writeTestConfig(t, tc.yaml))
			require.NoError(t, err)
			require.Equal(t, tc.want, cfg.User.PasswordLoginEnabled())
			require.NoError(t, cfg.Save())
			reloaded, _, err := LoadConfig(path)
			require.NoError(t, err)
			require.Equal(t, tc.want, reloaded.User.PasswordLoginEnabled())
		})
	}
}
