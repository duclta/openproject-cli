package configuration

import (
	"os"
	"reflect"
	"strings"
	"testing"

	keyring "github.com/zalando/go-keyring"
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
	keyring.MockInit()
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

	storedFile, err := os.ReadFile(configFile())
	if err != nil {
		t.Fatalf("read stored config file: %v", err)
	}

	storedText := string(storedFile)
	if strings.Contains(storedText, "secret") || strings.Contains(storedText, "csrf-token") || strings.Contains(storedText, "cookie-value") {
		t.Fatalf("expected session secrets to be scrubbed from config file: %s", storedText)
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

	expectedConfig := AuthConfig{
		Host:     "https://example.com",
		AuthType: AuthTypeSession,
		Username: "alice",
		Cookies: []Cookie{{
			Name:  "session",
			Value: "cookie-value",
			Path:  "/",
		}},
	}

	if !reflect.DeepEqual(readConfig, expectedConfig) {
		t.Fatalf("unexpected config: %#v", readConfig)
	}
}

func TestReadAuthConfig_MigratesStructuredSessionSecrets(t *testing.T) {
	keyring.MockInit()
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())

	legacyConfig := []byte(`{
	  "host": "https://example.com",
	  "authType": "session",
	  "username": "alice",
	  "password": "secret",
	  "csrfToken": "csrf-token",
	  "cookies": [
	    {
	      "name": "session",
	      "value": "cookie-value",
	      "path": "/"
	    }
	  ]
	}`)

	err := ensureConfigDir()
	if err != nil {
		t.Fatalf("ensure config dir: %v", err)
	}

	err = os.WriteFile(configFile(), legacyConfig, 0600)
	if err != nil {
		t.Fatalf("write legacy config: %v", err)
	}

	config, err := ReadAuthConfig()
	if err != nil {
		t.Fatalf("read auth config: %v", err)
	}

	expectedConfig := AuthConfig{
		Host:     "https://example.com",
		AuthType: AuthTypeSession,
		Username: "alice",
		Cookies: []Cookie{{
			Name:  "session",
			Value: "cookie-value",
			Path:  "/",
		}},
	}

	if !reflect.DeepEqual(config, expectedConfig) {
		t.Fatalf("unexpected config: %#v", config)
	}

	storedFile, err := os.ReadFile(configFile())
	if err != nil {
		t.Fatalf("read migrated config file: %v", err)
	}

	storedText := string(storedFile)
	if strings.Contains(storedText, "secret") || strings.Contains(storedText, "csrf-token") || strings.Contains(storedText, "cookie-value") {
		t.Fatalf("expected migrated config file to be scrubbed: %s", storedText)
	}

	migratedConfig, err := ReadAuthConfig()
	if err != nil {
		t.Fatalf("read migrated auth config again: %v", err)
	}

	if !reflect.DeepEqual(migratedConfig, expectedConfig) {
		t.Fatalf("unexpected migrated config: %#v", migratedConfig)
	}
}
