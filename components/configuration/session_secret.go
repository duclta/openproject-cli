package configuration

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/opf/openproject-cli/components/errors"
	keyring "github.com/zalando/go-keyring"
)

const keyringServiceName = "openproject-cli"

type storedSessionSecret struct {
	Cookies []Cookie `json:"cookies,omitempty"`
}

func hasPlaintextSessionSecrets(config AuthConfig) bool {
	return config.UsesSessionAuth() && (len(config.Password) > 0 || len(config.Cookies) > 0 || len(config.CSRFToken) > 0)
}

func storeSessionSecret(config AuthConfig) error {
	account, ok := sessionSecretAccount(config)
	if !ok || len(config.Cookies) == 0 {
		return nil
	}

	bytes, err := json.Marshal(storedSessionSecret{Cookies: config.Cookies})
	if err != nil {
		return err
	}

	return keyring.Set(keyringServiceName, account, string(bytes))
}

func loadSessionSecret(config AuthConfig) ([]Cookie, error) {
	account, ok := sessionSecretAccount(config)
	if !ok {
		return nil, nil
	}

	secret, err := keyring.Get(keyringServiceName, account)
	if err != nil {
		if err == keyring.ErrNotFound {
			return nil, nil
		}

		return nil, errors.Custom(fmt.Sprintf("Could not load stored session secret from OS keychain: %v", err))
	}

	if len(strings.TrimSpace(secret)) == 0 {
		return nil, nil
	}

	var storedSecret storedSessionSecret
	if err := json.Unmarshal([]byte(secret), &storedSecret); err != nil {
		return nil, errors.Custom("Stored session secret is invalid. Run `op login` again.")
	}

	return storedSecret.Cookies, nil
}

func sessionSecretAccount(config AuthConfig) (string, bool) {
	host := strings.TrimSpace(config.Host)
	username := strings.TrimSpace(config.Username)
	if !config.UsesSessionAuth() || len(host) == 0 || len(username) == 0 {
		return "", false
	}

	return fmt.Sprintf("session|%s|%s", host, username), true
}
