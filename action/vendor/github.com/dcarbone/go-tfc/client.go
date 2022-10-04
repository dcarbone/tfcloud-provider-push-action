package tfc

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/hashicorp/go-cleanhttp"
)

const (
	DefaultAddress = "https://app.terraform.io"

	headerAccept             = "Accept"
	headerAuthorization      = "Authorization"
	headerContentDisposition = "Content-Disposition"
	headerContentType        = "Content-Type"

	applicationVNDAPIJSON = "application/vnd.api+json"
	applicationJSON       = "application/json"
	binaryOctetStream     = "binary/octet-stream"
	attachmentFilenameFmt = "attachment; filename=%q"

	pathAPI               = "api"
	pathOrganizations     = "organizations"
	pathPlatforms         = "platforms"
	pathRegistryModules   = "registry-modules"
	pathRegistryProviders = "registry-providers"
	pathVersions          = "versions"
	pathV2                = "v2"
)

type Config struct {
	// Address [required] (default: https://app.terraform.io)
	//
	// Address, including scheme and port (if not well-known) of C
	Address string

	// HTTPClient [required] (default: basic http client with connection pooling)
	//
	// HTTP client instance to use
	HTTPClient *http.Client
}

type ConfigOption func(*Config)

type clientMiddleware struct {
	addr string
	hc   *http.Client
}

func newClientMiddleware(cfg *Config) (*clientMiddleware, error) {
	tm := new(clientMiddleware)
	tm.addr = strings.Trim(cfg.Address, "/")
	tm.hc = cleanhttp.DefaultPooledClient()

	return tm, nil
}

// copy does a shallow copy of the middleware, overriding the http client with the one provided.  used by the
// UploadsClient.
func (tm *clientMiddleware) copy(hc *http.Client) *clientMiddleware {
	tm2 := new(clientMiddleware)
	*tm2 = *tm
	tm2.hc = hc
	return tm2
}

func (tm *clientMiddleware) buildURL(route string) string {
	return fmt.Sprintf("%s/%s", tm.addr, route)
}

func (tm *clientMiddleware) buildRequest(ctx context.Context, method, routePath string, query url.Values, body interface{}) (*http.Request, error) {
	var bodyRdr io.Reader

	// handle different incoming body types
	if body != nil {
		switch body.(type) {
		case io.Reader:
			bodyRdr = body.(io.Reader)
		case []byte:
			bodyRdr = bytes.NewBuffer(body.([]byte))
		default:
			if b, err := json.Marshal(body); err != nil {
				return nil, fmt.Errorf("error marshalling body: %w", err)
			} else {
				bodyRdr = bytes.NewBuffer(b)
			}
		}
	}

	compiledURL := tm.buildURL(routePath)
	if len(query) > 0 {
		compiledURL = fmt.Sprintf("%s?%s", compiledURL, query.Encode())
	}

	r, err := http.NewRequestWithContext(ctx, method, compiledURL, bodyRdr)

	return r, err
}

func (tm *clientMiddleware) do(r *http.Request) (*http.Response, error) {
	// todo: this abstraction is here as i plan to eventually move additional logic here.
	return tm.hc.Do(r)
}

type Client struct {
	m *clientMiddleware
}

// NewClient constructs a new client instance based upon the provided configuration.  Specific-intent clients are
// chained from this base client instance.
func NewClient(cfg *Config, opts ...ConfigOption) (*Client, error) {
	var (
		err error

		tc = new(Client)
	)

	conf := compileConfig(cfg, opts...)

	if tc.m, err = newClientMiddleware(conf); err != nil {
		return nil, err
	}

	return tc, nil
}
