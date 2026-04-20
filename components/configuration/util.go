package configuration

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/opf/openproject-cli/components/common"
	"github.com/opf/openproject-cli/components/errors"
)

const (
	envHost        = "OP_CLI_HOST"
	envToken       = "OP_CLI_TOKEN"
	configDirName  = "openproject"
	configFileName = "config"
)

func WriteConfigFile(host, token string) error {
	return WriteAuthConfig(AuthConfig{Host: host, AuthType: AuthTypeAPIToken, Token: token})
}

func WriteAuthConfig(config AuthConfig) error {
	err := ensureConfigDir()
	if err != nil {
		return err
	}

	bytes, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(configFile(), bytes, 0600)
}

func ReadConfig() (host, token string, err error) {
	config, err := ReadAuthConfig()
	if err != nil {
		return "", "", err
	}

	return config.Host, config.Token, nil
}

func ReadAuthConfig() (AuthConfig, error) {
	err := ensureConfigDir()
	if err != nil {
		return AuthConfig{}, err
	}

	ok, h, t := readEnvironment()
	if ok {
		return AuthConfig{Host: h, AuthType: AuthTypeAPIToken, Token: t}, nil
	}

	file, err := os.ReadFile(configFile())
	if os.IsNotExist(err) {
		// Empty config file is no error,
		// user just has to run login command first
		return AuthConfig{}, nil
	}
	if err != nil {
		return AuthConfig{}, err
	}

	content := common.SanitizeLineBreaks(string(file))
	if len(strings.TrimSpace(content)) == 0 {
		return AuthConfig{}, nil
	}

	var config AuthConfig
	if err := json.Unmarshal(file, &config); err == nil && config.HasHost() {
		return config, nil
	}

	parts := strings.Fields(content)
	if len(parts) != 2 {
		return AuthConfig{}, errors.Custom(fmt.Sprintf("Invalid config file at %s. Please remove the file and run `op login` again.", configFile()))
	}

	return AuthConfig{Host: parts[0], AuthType: AuthTypeAPIToken, Token: parts[1]}, nil
}

func readEnvironment() (ok bool, host, token string) {
	host, hasHost := os.LookupEnv(envHost)
	token, hasToken := os.LookupEnv(envToken)
	ok = hasHost && hasToken

	return
}

func ensureConfigDir() error {
	if _, err := os.Stat(configFileDir()); os.IsNotExist(err) {
		err = os.MkdirAll(configFileDir(), 0700)
		if err != nil {
			return err
		}
	}

	return nil
}

func configFile() string {
	return filepath.Join(configFileDir(), configFileName)
}

func configFileDir() string {
	xdgConfigDir, present := os.LookupEnv("XDG_CONFIG_HOME")
	if present {
		return filepath.Join(xdgConfigDir, configDirName)
	}

	return filepath.Join(homeDir(), ".config", configDirName)
}

func homeDir() string {
	if home, ok := os.LookupEnv("HOME"); ok {
		return home
	}

	// On Windows `$HOME` is not set per default, but it is
	// constructed from `$HOMEDRIVE` and `$HOMEPATH`.
	return filepath.Join(os.Getenv("HOMEDRIVE"), os.Getenv("HOMEPATH"))
}
