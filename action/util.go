package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
)

func setBearerToken(req *http.Request, token string) {
	const f = "Bearer %s"
	req.Header.Set(headerAuthorization, fmt.Sprintf(f, token))
}

func buildPath(parts ...interface{}) string {
	out := ""
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
			out = fmt.Sprintf("%s/%s", out, v)
		}
	}
	return out
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
	return generateUnexpectedResponseCodeError(resp, httpCode)
}

func generateUnexpectedResponseCodeError(resp *http.Response, expected int) error {
	var buf bytes.Buffer
	_, _ = io.Copy(&buf, resp.Body)
	drainReader(resp.Body)

	// attempt to unmarshal api response error.  don't particularly care if we can't.
	tfcErr := new(TFCloudAPIError)
	_ = json.Unmarshal(buf.Bytes(), tfcErr)

	// get raw body bytes, just in case...
	trimmed := strings.TrimSpace(string(buf.Bytes()))

	return StatusError{
		ExpectedCode: expected,
		ActualCode:   resp.StatusCode,
		Body:         trimmed,
		TFCloudError: *tfcErr,
	}
}

// handleResponse is a helper func to reduce amount of repeated code in each api call.  will eventually
// be expanded upon.
func handleResponse(resp *http.Response, respErr error, out interface{}, expectedCode int) error {
	var err error
	if resp != nil {
		defer drainReader(resp.Body)
	}
	if respErr != nil {
		return respErr
	}
	if err = requireHTTPCodes(resp, expectedCode); err != nil {
		return err
	}
	if out != nil {
		if err = json.NewDecoder(resp.Body).Decode(out); err != nil {
			return nil
		}
	}
	return nil
}
