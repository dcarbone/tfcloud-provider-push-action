package tfc

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"path"
	"reflect"
	"strconv"
	"strings"

	"github.com/hashicorp/go-cleanhttp"
)

func applyConfigOpts(cf *Config, opts ...ConfigOption) {
	for _, opt := range opts {
		opt(cf)
	}
}

func compileConfig(userConf *Config, opts ...ConfigOption) *Config {
	actual := new(Config)
	if userConf != nil {
		*actual = *userConf
	}
	applyConfigOpts(actual, opts...)
	if actual.Address == "" {
		actual.Address = DefaultAddress
	}
	if actual.HTTPClient == nil {
		actual.HTTPClient = cleanhttp.DefaultPooledClient()
	}
	return actual
}

func setBearerToken(req *http.Request, token string) {
	const f = "Bearer %s"
	req.Header.Set(headerAuthorization, fmt.Sprintf(f, token))
}

func buildRoute(parts ...interface{}) string {
	partStrs := make([]string, 0)
	var v string
	for _, p := range parts {
		v = ""
		switch p.(type) {
		case string:
			v = p.(string)
		case int:
			v = strconv.Itoa(p.(int))
		case *int:
			if p.(*int) != nil {
				v = strconv.Itoa(*p.(*int))
			}
		default:
			panic(fmt.Sprintf("no path part handler defined for type \"%T\" (%[1]v)", p))
		}
		if v != "" {
			partStrs = append(partStrs, v)
		}
	}
	return path.Join(partStrs...)
}

func drainReader(r io.Reader) {
	if r == nil {
		return
	}
	_, _ = io.Copy(ioutil.Discard, r)
	if rc, ok := r.(io.Closer); ok {
		_ = rc.Close()
	}
}

func requireHTTPCodes(resp *http.Response, httpCode int) error {
	if resp.StatusCode == httpCode {
		return nil
	}
	return createUnexpectedResponseCodeError(resp, httpCode)
}

func createUnexpectedResponseCodeError(resp *http.Response, expected int) error {
	var buf bytes.Buffer
	_, _ = io.Copy(&buf, resp.Body)
	defer drainReader(resp.Body)

	// attempt to unmarshal api response error.  don't particularly care if we can't.
	apiErr := new(APIError)
	_ = json.Unmarshal(buf.Bytes(), apiErr)

	// get raw body bytes, just in case...
	trimmed := strings.TrimSpace(string(buf.Bytes()))

	return &StatusError{
		ExpectedCode: expected,
		ActualCode:   resp.StatusCode,
		Body:         trimmed,
		CloudError:   *apiErr,
	}
}

func createRequestExecutionError(req *http.Request, err error) error {
	return fmt.Errorf("error executing %s %q: %w", req.Method, req.URL, err)
}

// handleResponse is a helper func to reduce amount of repeated code in each api call.  will eventually
// be expanded upon.
func handleResponse(req *http.Request, resp *http.Response, out interface{}, expectedCode int) error {
	if resp != nil {
		defer drainReader(resp.Body)
	}
	if err := requireHTTPCodes(resp, expectedCode); err != nil {
		return createRequestExecutionError(req, err)
	}
	if out != nil {
		// ensure we've got a pointer
		if reflect.Ptr != reflect.TypeOf(out).Kind() {
			panic(fmt.Sprintf("Must provided pointer to handleResponse, saw %T", out))
		}
		// attempt decode
		if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
			return createRequestExecutionError(req, fmt.Errorf("error unmarshalling response into %T: %w", out, err))
		}
	}
	return nil
}
