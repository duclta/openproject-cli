package requests

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/opf/openproject-cli/components/configuration"
	"github.com/opf/openproject-cli/components/printer"
	keyring "github.com/zalando/go-keyring"
)

func TestDo_ReauthenticatesSessionOnUnauthorized(t *testing.T) {
	printer.Init(&printer.TestingPrinter{})
	keyring.MockInit()
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())

	server, state := newSessionTestServer(t)
	defer server.Close()

	hostURL, err := url.Parse(server.URL)
	if err != nil {
		t.Fatalf("parse server url: %v", err)
	}

	startingAuth := configuration.AuthConfig{
		Host:     server.URL,
		AuthType: configuration.AuthTypeSession,
		Username: state.username,
		Cookies: []configuration.Cookie{{
			Name:  sessionCookieName,
			Value: "stale-session",
			Path:  "/",
		}, {
			Name:  autologinCookieName,
			Value: freshAutologinValue,
			Path:  "/",
		}},
	}

	err = configuration.WriteAuthConfig(startingAuth)
	if err != nil {
		t.Fatalf("write stored auth config: %v", err)
	}

	storedAuth, err := configuration.ReadAuthConfig()
	if err != nil {
		t.Fatalf("read stored auth config: %v", err)
	}

	Init(hostURL, storedAuth, false)

	body, err := Do("GET", "/api/v3/projects", nil, nil)
	if err != nil {
		t.Fatalf("get projects: %v", err)
	}

	if !strings.Contains(string(body), `"_type":"Collection"`) {
		t.Fatalf("unexpected response body: %s", body)
	}

	if state.loginCount != 0 {
		t.Fatalf("expected cookie-based refresh without password login, got %d logins", state.loginCount)
	}

	storedConfig, err := configuration.ReadAuthConfig()
	if err != nil {
		t.Fatalf("read stored config: %v", err)
	}

	if storedConfig.CSRFToken != "" {
		t.Fatalf("expected csrf token to stay out of persisted config, got %s", storedConfig.CSRFToken)
	}

	if cookieValue(storedConfig.Cookies, sessionCookieName) != freshSessionValue {
		t.Fatalf("unexpected stored cookies: %#v", storedConfig.Cookies)
	}

	if cookieValue(storedConfig.Cookies, autologinCookieName) != freshAutologinValue {
		t.Fatalf("unexpected stored autologin cookie: %#v", storedConfig.Cookies)
	}
}

func TestDo_RewindsRequestBodyAfterSessionRefresh(t *testing.T) {
	printer.Init(&printer.TestingPrinter{})
	keyring.MockInit()
	t.Setenv("XDG_CONFIG_HOME", t.TempDir())

	server, state := newSessionTestServer(t)
	defer server.Close()

	hostURL, err := url.Parse(server.URL)
	if err != nil {
		t.Fatalf("parse server url: %v", err)
	}

	payload := []byte(`{"subject":"Updated via session"}`)

	startingAuth := configuration.AuthConfig{
		Host:     server.URL,
		AuthType: configuration.AuthTypeSession,
		Username: state.username,
		Cookies: []configuration.Cookie{{
			Name:  sessionCookieName,
			Value: "stale-session",
			Path:  "/",
		}, {
			Name:  autologinCookieName,
			Value: freshAutologinValue,
			Path:  "/",
		}},
	}

	err = configuration.WriteAuthConfig(startingAuth)
	if err != nil {
		t.Fatalf("write stored auth config: %v", err)
	}

	storedAuth, err := configuration.ReadAuthConfig()
	if err != nil {
		t.Fatalf("read stored auth config: %v", err)
	}

	Init(hostURL, storedAuth, false)

	body, err := Do("PATCH", "/api/v3/work_packages/1", nil, &RequestData{
		ContentType: "application/json",
		Body:        bytes.NewReader(payload),
	})
	if err != nil {
		t.Fatalf("patch work package: %v", err)
	}

	if string(body) != string(payload) {
		t.Fatalf("unexpected response body: %s", body)
	}

	if state.patchBody != string(payload) {
		t.Fatalf("unexpected patch body: %s", state.patchBody)
	}

	if state.patchCSRFToken != freshCSRFToken {
		t.Fatalf("unexpected csrf header: %s", state.patchCSRFToken)
	}

	if state.loginCount != 0 {
		t.Fatalf("expected cookie-based refresh without password login, got %d logins", state.loginCount)
	}
}

const (
	sessionCookieName   = "session"
	autologinCookieName = "autologin"
	freshSessionValue   = "fresh-session"
	freshAutologinValue = "fresh-autologin"
	freshCSRFToken      = "fresh-csrf"
	loginFormToken      = "login-form-token"
)

type sessionTestState struct {
	username       string
	password       string
	serverURL      string
	loginCount     int
	patchBody      string
	patchCSRFToken string
}

func newSessionTestServer(t *testing.T) (*httptest.Server, *sessionTestState) {
	t.Helper()

	state := &sessionTestState{
		username: "alice",
		password: "secret",
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/login":
			handleLoginRequest(w, r, state)
		case "/":
			if hasFreshAutologin(r) {
				http.SetCookie(w, &http.Cookie{Name: sessionCookieName, Value: freshSessionValue, Path: "/"})
			}
			fmt.Fprintf(w, `<html><head><meta name="csrf-token" content="%s"></head></html>`, freshCSRFToken)
		case "/api/v3/users/me":
			if !hasFreshSession(r) {
				_, _ = io.WriteString(w, `{"id":0,"name":"Anonymous"}`)
				return
			}

			w.Header().Set("Content-Type", "application/json")
			_, _ = io.WriteString(w, `{"id":1,"name":"Alice"}`)
		case "/api/v3/projects":
			if r.Header.Get("X-Requested-With") != "XMLHttpRequest" || !hasFreshSession(r) {
				w.WriteHeader(http.StatusUnauthorized)
				_, _ = io.WriteString(w, `{"error":"expired"}`)
				return
			}

			w.Header().Set("Content-Type", "application/json")
			_, _ = io.WriteString(w, `{"_type":"Collection"}`)
		case "/api/v3/work_packages/1":
			if r.Header.Get("X-Requested-With") != "XMLHttpRequest" || !hasFreshSession(r) {
				w.WriteHeader(http.StatusUnauthorized)
				_, _ = io.WriteString(w, `{"error":"expired"}`)
				return
			}

			state.patchCSRFToken = r.Header.Get("X-CSRF-Token")
			body, err := io.ReadAll(r.Body)
			if err != nil {
				t.Fatalf("read patch body: %v", err)
			}
			state.patchBody = string(body)

			if state.patchCSRFToken != freshCSRFToken {
				w.WriteHeader(http.StatusUnprocessableEntity)
				_, _ = io.WriteString(w, `invalid authenticity token`)
				return
			}

			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write(body)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))

	state.serverURL = server.URL
	return server, state
}

func handleLoginRequest(w http.ResponseWriter, r *http.Request, state *sessionTestState) {
	switch r.Method {
	case http.MethodGet:
		http.SetCookie(w, &http.Cookie{Name: sessionCookieName, Value: "prelogin", Path: "/"})
		fmt.Fprintf(w, `<html><head><meta name="csrf-token" content="login-csrf"></head><body><form action="/login" method="post"><input type="hidden" name="authenticity_token" value="%s" /><input type="hidden" name="back_url" value="%s/login" /></form></body></html>`, loginFormToken, state.serverURL)
	case http.MethodPost:
		if err := r.ParseForm(); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		if r.Form.Get("username") != state.username || r.Form.Get("password") != state.password || r.Form.Get("authenticity_token") != loginFormToken {
			_, _ = io.WriteString(w, `<html><head><meta name="csrf-token" content="login-csrf"></head><body>login failed</body></html>`)
			return
		}

		state.loginCount++
		http.SetCookie(w, &http.Cookie{Name: sessionCookieName, Value: freshSessionValue, Path: "/"})
		http.SetCookie(w, &http.Cookie{Name: autologinCookieName, Value: freshAutologinValue, Path: "/"})
		http.Redirect(w, r, "/", http.StatusFound)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func hasFreshSession(r *http.Request) bool {
	cookie, err := r.Cookie(sessionCookieName)
	if err != nil {
		return false
	}

	return cookie.Value == freshSessionValue
}

func hasFreshAutologin(r *http.Request) bool {
	cookie, err := r.Cookie(autologinCookieName)
	if err != nil {
		return false
	}

	return cookie.Value == freshAutologinValue
}

func cookieValue(cookies []configuration.Cookie, name string) string {
	for _, cookie := range cookies {
		if cookie.Name == name {
			return cookie.Value
		}
	}

	return ""
}
