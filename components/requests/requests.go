package requests

import (
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"

	"github.com/opf/openproject-cli/components/configuration"
	"github.com/opf/openproject-cli/components/errors"
	"github.com/opf/openproject-cli/components/printer"
)

var client *http.Client
var host *url.URL
var auth configuration.AuthConfig
var verbose bool

type RequestData struct {
	ContentType string
	Body        io.Reader
}

func Init(hostUrl *url.URL, authConfig configuration.AuthConfig, verboseFlag bool) {
	verbose = verboseFlag
	host = cloneURL(hostUrl)
	auth = authConfig

	httpClient := &http.Client{}
	if auth.UsesSessionAuth() {
		jar, err := cookiejar.New(nil)
		if err != nil {
			printer.Error(err)
			httpClient = &http.Client{}
		} else {
			httpClient.Jar = jar
			setSessionCookies(jar, host, auth.Cookies)
		}
	}

	client = httpClient
}

func Get(path string, query *Query) (responseBody []byte, err error) {
	loadingFunc := func() ([]byte, error) { return Do("GET", path, query, nil) }

	return printer.WithSpinner(loadingFunc)
}

func Post(path string, requestData *RequestData) (responseBody []byte, err error) {
	loadingFunc := func() ([]byte, error) { return Do("POST", path, nil, requestData) }

	return printer.WithSpinner(loadingFunc)
}

func Patch(path string, requestBody *RequestData) (responseBody []byte, err error) {
	loadingFunc := func() ([]byte, error) { return Do("PATCH", path, nil, requestBody) }

	return printer.WithSpinner(loadingFunc)
}

func Do(method string, path string, query *Query, requestData *RequestData) (responseBody []byte, err error) {
	return do(method, path, query, requestData, true)
}

func Probe(path string) (statusCode int, header http.Header, body []byte, err error) {
	if client == nil || hostUnitialised() {
		return 0, nil, nil, errors.Custom("Cannot execute requests without initializing request client first. Run `op login`")
	}

	request, err := buildRequest("GET", path, nil, nil)
	if err != nil {
		return 0, nil, nil, err
	}

	resp, responseBody, err := execute(request)
	if err != nil {
		return 0, nil, nil, err
	}

	return resp.StatusCode, resp.Header, responseBody, nil
}

func do(method string, path string, query *Query, requestData *RequestData, allowRetry bool) (responseBody []byte, err error) {
	printer.Debug(verbose, "Building HTTP request:")
	printer.Debug(verbose, fmt.Sprintf("\twith Method: %s", method))
	printer.Debug(verbose, fmt.Sprintf("\twith Path: %s", path))
	printer.Debug(verbose, fmt.Sprintf("\twith Query: %+v", query))
	printer.Debug(verbose, fmt.Sprintf("\twith Body: %+v", requestData))

	if client == nil || hostUnitialised() {
		return nil, errors.Custom("Cannot execute requests without initializing request client first. Run `op login`")
	}

	request, err := buildRequest(method, path, query, requestData)
	if err != nil {
		return nil, err
	}

	resp, response, err := execute(request)
	if err != nil {
		return nil, err
	}

	printer.Debug(verbose, fmt.Sprintf("Received response:\n%s", response))

	if allowRetry && shouldReAuthenticate(method, resp.StatusCode, response) {
		printer.Debug(verbose, "Authentication expired, re-establishing session ...")
		err = reAuthenticate()
		if err != nil {
			return nil, err
		}

		return do(method, path, query, requestData, false)
	}

	if !isSuccess(resp.StatusCode) {
		return nil, errors.NewResponseError(resp.StatusCode, response)
	}

	return response, nil
}

func buildRequest(method string, path string, query *Query, requestData *RequestData) (*http.Request, error) {
	err := rewindRequestBody(requestData)
	if err != nil {
		return nil, err
	}

	requestUrl := *host
	requestUrl.Path += path
	if query != nil {
		requestUrl.RawQuery = query.String()
	}

	var body io.Reader
	if requestData != nil {
		body = requestData.Body
	}

	request, err := http.NewRequest(strings.ToUpper(method), requestUrl.String(), body)
	if err != nil {
		return nil, err
	}

	if requestData != nil {
		request.Header.Add("Content-Type", requestData.ContentType)
	}

	applyAuthentication(request)

	return request, nil
}

func applyAuthentication(request *http.Request) {
	if auth.UsesAPIToken() {
		request.SetBasicAuth("apikey", auth.Token)
		return
	}

	if auth.UsesSessionAuth() {
		request.Header.Set("X-Requested-With", "XMLHttpRequest")
		if requiresCSRFToken(request.Method) && len(auth.CSRFToken) > 0 {
			request.Header.Set("X-CSRF-Token", auth.CSRFToken)
		}
	}
}

func execute(request *http.Request) (*http.Response, []byte, error) {
	printer.Debug(verbose, fmt.Sprintf("Running HTTP request %s %s", request.Method, request.URL))

	resp, err := client.Do(request)
	if err != nil {
		return nil, nil, err
	}

	defer func() { _ = resp.Body.Close() }()
	response, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, nil, err
	}

	return resp, response, nil
}

func shouldReAuthenticate(method string, statusCode int, response []byte) bool {
	if !auth.UsesSessionAuth() || len(auth.Username) == 0 || len(auth.Password) == 0 {
		return false
	}

	if statusCode == http.StatusUnauthorized {
		return true
	}

	if statusCode == http.StatusUnprocessableEntity && requiresCSRFToken(method) {
		lowerBody := strings.ToLower(string(response))
		return strings.Contains(lowerBody, "csrf") || strings.Contains(lowerBody, "authenticity")
	}

	return false
}

func reAuthenticate() error {
	refreshedAuth, err := AuthenticateSession(host, auth.Username, auth.Password, verbose)
	if err != nil {
		return err
	}

	Init(host, refreshedAuth, verbose)
	return configuration.WriteAuthConfig(refreshedAuth)
}

func rewindRequestBody(requestData *RequestData) error {
	if requestData == nil || requestData.Body == nil {
		return nil
	}

	seeker, ok := requestData.Body.(io.Seeker)
	if !ok {
		return nil
	}

	_, err := seeker.Seek(0, io.SeekStart)
	return err
}

func requiresCSRFToken(method string) bool {
	switch strings.ToUpper(method) {
	case "GET", "HEAD", "OPTIONS":
		return false
	default:
		return true
	}
}

func isSuccess(code int) bool {
	return code >= 200 && code <= 299
}

func hostUnitialised() bool {
	return host == nil || len(host.Scheme) == 0 || len(host.Host) == 0
}

func cloneURL(input *url.URL) *url.URL {
	if input == nil {
		return &url.URL{}
	}

	clone := *input
	return &clone
}
