package configuration

import "net/http"

type AuthType string

const (
	AuthTypeAPIToken AuthType = "api_token"
	AuthTypeSession  AuthType = "session"
)

type Cookie struct {
	Name     string `json:"name"`
	Value    string `json:"value"`
	Path     string `json:"path,omitempty"`
	Domain   string `json:"domain,omitempty"`
	Secure   bool   `json:"secure,omitempty"`
	HttpOnly bool   `json:"httpOnly,omitempty"`
}

type AuthConfig struct {
	Host      string   `json:"host,omitempty"`
	AuthType  AuthType `json:"authType,omitempty"`
	Token     string   `json:"token,omitempty"`
	Username  string   `json:"username,omitempty"`
	Password  string   `json:"password,omitempty"`
	CSRFToken string   `json:"csrfToken,omitempty"`
	Cookies   []Cookie `json:"cookies,omitempty"`
}

func (config AuthConfig) UsesSessionAuth() bool {
	return config.AuthType == AuthTypeSession
}

func (config AuthConfig) UsesAPIToken() bool {
	return config.AuthType == AuthTypeAPIToken && len(config.Token) > 0
}

func (config AuthConfig) HasHost() bool {
	return len(config.Host) > 0
}

func (cookie Cookie) ToHTTPCookie() *http.Cookie {
	return &http.Cookie{
		Name:     cookie.Name,
		Value:    cookie.Value,
		Path:     cookie.Path,
		Domain:   cookie.Domain,
		Secure:   cookie.Secure,
		HttpOnly: cookie.HttpOnly,
	}
}

func CookieFromHTTP(cookie *http.Cookie) Cookie {
	return Cookie{
		Name:     cookie.Name,
		Value:    cookie.Value,
		Path:     cookie.Path,
		Domain:   cookie.Domain,
		Secure:   cookie.Secure,
		HttpOnly: cookie.HttpOnly,
	}
}
