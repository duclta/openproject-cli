package cmd

import (
	"bufio"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"golang.org/x/term"

	"github.com/opf/openproject-cli/components/common"
	"github.com/opf/openproject-cli/components/configuration"
	"github.com/opf/openproject-cli/components/parser"
	"github.com/opf/openproject-cli/components/paths"
	"github.com/opf/openproject-cli/components/printer"
	"github.com/opf/openproject-cli/components/requests"
	"github.com/opf/openproject-cli/components/resources/users"
	"github.com/opf/openproject-cli/dtos"
)

type loginMethod string

const (
	loginMethodSession loginMethod = "session"
	loginMethodToken   loginMethod = "token"
)

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Authenticates the user against an OpenProject instance",
	Long: `Enables the login flow, which enables the user to use
this tool for a specific OpenProject instance. The login
needs the host URL of the OpenProject instance and either a
generated API token or a username/password pair.`,
	Run: login,
}

const (
	urlInputError        = "There was a problem parsing the input. Please try again and put in a valid URL."
	missingSchemeError   = "URL scheme is missing, please define a complete URL."
	noOpInstanceError    = "URL does not point to a valid OpenProject instance."
	tokenInputError      = "There was a problem parsing the token input. Please try again."
	usernameInputError   = "There was a problem parsing the username input. Please try again."
	passwordInputError   = "There was a problem parsing the password input. Please try again."
	authMethodInputError = "Unknown authentication method. Please choose `session` or `token`."
)

func login(_ *cobra.Command, _ []string) {
	var hostUrl *url.URL

	for {
		printer.Debug(Verbose, "Parsing host URL ...")
		printer.Input("OpenProject host URL: ")

		ok, msg, host := parseHostUrl()
		if !ok {
			printer.ErrorText(msg)
			continue
		}

		printer.Debug(Verbose, "Initializing requests client ...")
		requests.Init(host, configuration.AuthConfig{}, Verbose)
		ok = checkOpenProjectApi()
		if !ok {
			printer.ErrorText(noOpInstanceError)
			continue
		}

		hostUrl = host
		break
	}

	var authConfig configuration.AuthConfig
	for {
		printer.Input("Authentication method [session/token] (default: session): ")
		ok, method := requestLoginMethod()
		if !ok {
			printer.ErrorText(authMethodInputError)
			continue
		}

		switch method {
		case loginMethodToken:
			authConfig = loginWithAPIToken(hostUrl)
		default:
			authConfig = loginWithSession(hostUrl)
		}

		break
	}

	storeLoginData(authConfig)
}

func loginWithAPIToken(hostUrl *url.URL) configuration.AuthConfig {
	for {
		printer.Input(fmt.Sprintf("OpenProject API Token (Visit %s/my/access_tokens to generate one): ", hostUrl))
		ok, token := requestApiToken()
		if !ok {
			printer.ErrorText(tokenInputError)
			continue
		}

		authConfig := configuration.AuthConfig{
			Host:     hostUrl.String(),
			AuthType: configuration.AuthTypeAPIToken,
			Token:    common.SanitizeLineBreaks(token),
		}

		if validateAuthenticatedUser(hostUrl, authConfig) {
			return authConfig
		}
	}
}

func loginWithSession(hostUrl *url.URL) configuration.AuthConfig {
	for {
		printer.Input("OpenProject username or email: ")
		ok, username := requestLineInput()
		if !ok {
			printer.ErrorText(usernameInputError)
			continue
		}

		printer.Input("OpenProject password: ")
		ok, password := requestPassword()
		if !ok {
			printer.ErrorText(passwordInputError)
			continue
		}

		authConfig, err := requests.AuthenticateSession(hostUrl, common.SanitizeLineBreaks(username), password, Verbose)
		if err != nil {
			printer.Error(err)
			continue
		}

		if validateAuthenticatedUser(hostUrl, authConfig) {
			return authConfig
		}
	}
}

func validateAuthenticatedUser(hostUrl *url.URL, authConfig configuration.AuthConfig) bool {
	requests.Init(hostUrl, authConfig, Verbose)

	user, err := users.Me()
	if err != nil {
		printer.Error(err)
		return false
	}

	if user.Name == "Anonymous" {
		printer.ErrorText("No authenticated user returned.")
		return false
	}

	return true
}

func parseHostUrl() (ok bool, errMessage string, host *url.URL) {
	readOk, input := requestLineInput()
	if !readOk {
		return false, urlInputError, nil
	}

	printer.Debug(Verbose, fmt.Sprintf("Parsed input %q.", input))
	printer.Debug(Verbose, "Sanitizing input ...")

	input = common.SanitizeLineBreaks(input)
	input = strings.TrimSuffix(input, "/")

	printer.Debug(Verbose, fmt.Sprintf("Sanitized input '%s'.", input))
	printer.Debug(Verbose, "Parsing input as url ...")

	parsed, err := url.Parse(input)
	if err != nil {
		printer.Debug(Verbose, fmt.Sprintf("Error parsing url: %+v", err))
		return false, urlInputError, nil
	}

	printer.Debug(Verbose, fmt.Sprintf("Parsed url '%s'.", parsed))
	printer.Debug(Verbose, "Checking for http host and scheme ...")

	if parsed.Scheme == "" || parsed.Host == "" {
		return false, missingSchemeError, nil
	}

	printer.Debug(Verbose, "Parsing input successful, continuing with next steps.")
	return true, "", parsed
}

func checkOpenProjectApi() bool {
	printer.Debug(Verbose, "Fetching API root to check for instance configuration ...")

	statusCode, header, body, err := requests.Probe(paths.Root())
	if err != nil {
		printer.Debug(Verbose, fmt.Sprintf("Error probing OpenProject API: %+v", err))
		return false
	}

	// Public instance: standard check on the root resource
	if statusCode == http.StatusOK {
		c := parser.Parse[dtos.ConfigDto](body)
		return c.Type == "Root" && len(c.InstanceName) > 0
	}

	// Auth-required instance: detect OpenProject via the Link header added before authentication
	if statusCode == http.StatusUnauthorized {
		linkHeader := header.Get("Link")
		if strings.Contains(linkHeader, "/api/v3/openapi.json") {
			return true
		}
		// Fallback: check error body for OpenProject-specific error identifier
		return strings.Contains(string(body), "openproject-org")
	}

	return false
}

func requestApiToken() (ok bool, token string) {
	return requestLineInput()
}

func requestLoginMethod() (bool, loginMethod) {
	ok, input := requestLineInput()
	if !ok {
		return false, ""
	}

	switch strings.ToLower(strings.TrimSpace(input)) {
	case "", "session", "username", "username/password", "password":
		return true, loginMethodSession
	case "token", "api token", "api-token":
		return true, loginMethodToken
	default:
		return false, ""
	}
}

func requestLineInput() (bool, string) {
	reader := bufio.NewReader(os.Stdin)

	input, err := reader.ReadString('\n')
	if err != nil {
		printer.Debug(Verbose, fmt.Sprintf("Error reading string input: %+v", err))
		return false, ""
	}

	return true, input
}

func requestPassword() (bool, string) {
	password, err := term.ReadPassword(int(os.Stdin.Fd()))
	printer.Info("")
	if err != nil {
		printer.Debug(Verbose, fmt.Sprintf("Error reading password input: %+v", err))
		return false, ""
	}

	return true, string(password)
}

func storeLoginData(authConfig configuration.AuthConfig) {
	err := configuration.WriteAuthConfig(authConfig)
	if err != nil {
		printer.Error(err)
	}
}
