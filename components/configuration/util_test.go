package configuration

import (
	"os"
	"reflect"
	"testing"
)

func TestReadAuthConfig_LegacyTokenFormat(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())

	err := ensureConfigDir()
	if err != nil {
		t.Fatalf("ensure config dir: %v", err)
	}

	err = os.WriteFile(configFile(), []byte("https://example.com legacy-token"), 0600)
	if err != nil {
		t.Fatalf("write legacy config: %v", err)
	}

	config, err := ReadAuthConfig()
	if err != nil {
		t.Fatalf("read auth config: %v", err)
	}

	if config.Host != "https://example.com" {
		t.Fatalf("unexpected host: %s", config.Host)
	}

	if config.AuthType != AuthTypeAPIToken {
		t.Fatalf("unexpected auth type: %s", config.AuthType)
	}

	if config.Token != "legacy-token" {
		t.Fatalf("unexpected token: %s", config.Token)
	}
}

func TestWriteAuthConfig_RoundTripsStructuredConfig(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())

	writtenConfig := AuthConfig{
		Host:      "https://example.com",
		AuthType:  AuthTypeSession,
		Username:  "alice",
		Password:  "secret",
		CSRFToken: "csrf-token",
		Cookies: []Cookie{{
			Name:  "session",
			Value: "cookie-value",
			Path:  "/",
		}},
	}

	err := WriteAuthConfig(writtenConfig)
	if err != nil {
		t.Fatalf("write auth config: %v", err)
	}

	info, err := os.Stat(configFile())
	if err != nil {
		t.Fatalf("stat config file: %v", err)
	}

	if info.Mode().Perm() != 0600 {
		t.Fatalf("unexpected file mode: %o", info.Mode().Perm())
	}

	readConfig, err := ReadAuthConfig()
	if err != nil {
		t.Fatalf("read auth config: %v", err)
	}

	if !reflect.DeepEqual(readConfig, writtenConfig) {
		t.Fatalf("unexpected config: %#v", readConfig)
	}
}
