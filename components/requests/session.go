package requests

import (
	"fmt"
	"html"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"regexp"
	"strings"

	"github.com/opf/openproject-cli/components/configuration"
	"github.com/opf/openproject-cli/components/errors"
	"github.com/opf/openproject-cli/components/printer"
)

var (
	csrfMetaPattern          = regexp.MustCompile(`<meta name="csrf-token" content="([^"]+)"`)
	authenticityTokenPattern = regexp.MustCompile(`name="authenticity_token" value="([^"]+)"`)
	backURLPattern           = regexp.MustCompile(`name="back_url"[^>]*value="([^"]*)"`)
	loginErrorPattern        = regexp.MustCompile(`<p class="Banner-title"[^>]*>([^<]+)</p>`)
)

func AuthenticateSession(hostURL *url.URL, username, password string, verboseFlag bool) (configuration.AuthConfig, error) {
	printer.Debug(verboseFlag, "Starting web session authentication ...")

	httpClient, err := newSessionClient()
	if err != nil {
		return configuration.AuthConfig{}, err
	}

	loginPage, err := fetchPage(httpClient, resolvePath(hostURL, "/login"))
	if err != nil {
		return configuration.AuthConfig{}, err
	}

	authenticityToken, err := matchFirstGroup(authenticityTokenPattern, loginPage, "authenticity token")
	if err != nil {
		return configuration.AuthConfig{}, err
	}

	backURL := hostURL.String()
	if matchedBackURL, matchErr := matchFirstGroup(backURLPattern, loginPage, "back url"); matchErr == nil && len(matchedBackURL) > 0 {
		backURL = matchedBackURL
	}

	err = submitLogin(httpClient, hostURL, username, password, authenticityToken, backURL)
	if err != nil {
		return configuration.AuthConfig{}, err
	}

	err = verifyAuthenticatedSession(httpClient, hostURL)
	if err != nil {
		return configuration.AuthConfig{}, err
	}

	csrfToken, err := fetchCSRFToken(httpClient, hostURL)
	if err != nil {
		return configuration.AuthConfig{}, err
	}

	return configuration.AuthConfig{
		Host:      hostURL.String(),
		AuthType:  configuration.AuthTypeSession,
		Username:  username,
		Password:  password,
		CSRFToken: csrfToken,
		Cookies:   cookiesForHost(httpClient, hostURL),
	}, nil
}

func newSessionClient() (*http.Client, error) {
	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, err
	}

	return &http.Client{Jar: jar}, nil
}

func submitLogin(httpClient *http.Client, hostURL *url.URL, username, password, authenticityToken, backURL string) error {
	form := url.Values{}
	form.Set("username", username)
	form.Set("password", password)
	form.Set("authenticity_token", authenticityToken)
	form.Set("back_url", backURL)
	form.Set("autologin", "1")

	request, err := http.NewRequest("POST", resolvePath(hostURL, "/login").String(), strings.NewReader(form.Encode()))
	if err != nil {
		return err
	}

	request.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, body, err := executeWithClient(httpClient, request)
	if err != nil {
		return err
	}

	if resp.StatusCode >= 400 {
		return loginErrorFromResponse(resp.StatusCode, body)
	}

	return nil
}

func loginErrorFromResponse(status int, body []byte) error {
	if status == http.StatusUnprocessableEntity {
		if message, err := matchFirstGroup(loginErrorPattern, string(body), "login error"); err == nil {
			return errors.Custom(strings.TrimSpace(html.UnescapeString(message)))
		}
	}

	return errors.NewResponseError(status, body)
}

func verifyAuthenticatedSession(httpClient *http.Client, hostURL *url.URL) error {
	request, err := http.NewRequest("GET", resolvePath(hostURL, "/api/v3/users/me").String(), nil)
	if err != nil {
		return err
	}

	request.Header.Set("X-Requested-With", "XMLHttpRequest")

	resp, body, err := executeWithClient(httpClient, request)
	if err != nil {
		return err
	}

	if resp.StatusCode >= 400 {
		return errors.NewResponseError(resp.StatusCode, body)
	}

	if strings.Contains(strings.ToLower(string(body)), "anonymous") {
		return errors.Custom("Invalid username or password.")
	}

	return nil
}

func fetchCSRFToken(httpClient *http.Client, hostURL *url.URL) (string, error) {
	homePage, err := fetchPage(httpClient, resolvePath(hostURL, "/"))
	if err != nil {
		return "", err
	}

	return matchFirstGroup(csrfMetaPattern, homePage, "csrf token")
}

func fetchPage(httpClient *http.Client, target *url.URL) (string, error) {
	request, err := http.NewRequest("GET", target.String(), nil)
	if err != nil {
		return "", err
	}

	resp, body, err := executeWithClient(httpClient, request)
	if err != nil {
		return "", err
	}

	if resp.StatusCode >= 400 {
		return "", errors.NewResponseError(resp.StatusCode, body)
	}

	return string(body), nil
}

func executeWithClient(httpClient *http.Client, request *http.Request) (*http.Response, []byte, error) {
	resp, err := httpClient.Do(request)
	if err != nil {
		return nil, nil, err
	}

	defer func() { _ = resp.Body.Close() }()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, nil, err
	}

	return resp, body, nil
}

func resolvePath(hostURL *url.URL, path string) *url.URL {
	resolved := cloneURL(hostURL)
	resolved.Path = strings.TrimRight(resolved.Path, "/") + path
	resolved.RawQuery = ""
	resolved.Fragment = ""
	return resolved
}

func matchFirstGroup(pattern *regexp.Regexp, source string, valueName string) (string, error) {
	matches := pattern.FindStringSubmatch(source)
	if len(matches) != 2 {
		return "", errors.Custom(fmt.Sprintf("Could not extract %s from OpenProject login page.", valueName))
	}

	return matches[1], nil
}

func setSessionCookies(jar http.CookieJar, hostURL *url.URL, cookies []configuration.Cookie) {
	if jar == nil || hostURL == nil || len(cookies) == 0 {
		return
	}

	serialisedCookies := make([]*http.Cookie, 0, len(cookies))
	for _, cookie := range cookies {
		serialisedCookies = append(serialisedCookies, cookie.ToHTTPCookie())
	}

	jar.SetCookies(hostURL, serialisedCookies)
}

func cookiesForHost(httpClient *http.Client, hostURL *url.URL) []configuration.Cookie {
	if httpClient == nil || httpClient.Jar == nil || hostURL == nil {
		return nil
	}

	storedCookies := httpClient.Jar.Cookies(hostURL)
	serialisedCookies := make([]configuration.Cookie, 0, len(storedCookies))
	for _, cookie := range storedCookies {
		serialisedCookies = append(serialisedCookies, configuration.CookieFromHTTP(cookie))
	}

	return serialisedCookies
}
