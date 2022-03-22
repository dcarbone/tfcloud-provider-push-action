package tfc

import (
	"context"
	"net/http"
)

// ProviderClient interfaces with the Provider Registry in Terraform Cloud
type ProviderClient struct {
	m *clientMiddleware
}

// ProviderClient interfaces with the Provider Registry in Terraform Cloud
func (c *Client) ProviderClient() *ProviderClient {
	return &ProviderClient{m: c.m}
}

// CreateProviderVersion
//
// Executes: POST /api/v2/organizations/:organization_name/registry-providers/:registry_name/:namespace/:provider_name/versions
// Docs:     https://www.terraform.io/cloud-docs/api-docs/private-registry/provider-versions-platforms#create-a-provider-version
func (tc *ProviderClient) CreateProviderVersion(
	ctx context.Context,
	bearerToken,
	organizationName,
	registryName,
	namespace,
	providerName string,
	data CreateProviderVersionRequest,
) (*CreateProviderVersionResponse, error) {
	route := buildRoute(
		pathAPI,
		pathV2,
		pathOrganizations,
		organizationName,
		pathRegistryProviders,
		registryName,
		namespace,
		providerName,
		pathVersions,
	)
	req, err := tc.m.buildRequest(ctx, http.MethodPost, route, nil, data)
	if err != nil {
		return nil, err
	}
	setBearerToken(req, bearerToken)
	req.Header.Set(headerContentType, applicationVNDAPIJSON)
	req.Header.Set(headerAccept, applicationJSON)
	resp, err := tc.m.do(req)
	if err != nil {
		return nil, err
	}
	out := CreateProviderVersionResponse{}
	if err = handleResponse(req, resp, &out, http.StatusCreated); err != nil {
		return nil, err
	}
	return &out, nil
}

// CreateProviderVersionPlatform
//
// Executes: POST /api/v2/organizations/:organization_name/registry-providers/:registry_name/:namespace/:provider_name/versions/:provider_version/platforms
// Docs:     https://www.terraform.io/cloud-docs/api-docs/private-registry/provider-versions-platforms#create-a-provider-platform
func (tc *ProviderClient) CreateProviderVersionPlatform(
	ctx context.Context,
	bearerToken,
	organizationName,
	registryName,
	namespace,
	providerName,
	providerVersion string,
	data CreateProviderVersionPlatformRequest,
) (*CreateProviderVersionPlatformResponse, error) {
	route := buildRoute(
		pathAPI,
		pathV2,
		pathOrganizations,
		organizationName,
		pathRegistryProviders,
		registryName,
		namespace,
		providerName,
		pathVersions,
		providerVersion,
		pathPlatforms,
	)
	req, err := tc.m.buildRequest(ctx, http.MethodPost, route, nil, data)
	if err != nil {
		return nil, err
	}
	setBearerToken(req, bearerToken)
	req.Header.Set(headerContentType, applicationVNDAPIJSON)
	req.Header.Set(headerAccept, applicationJSON)
	resp, err := tc.m.do(req)
	if err != nil {
		return nil, err
	}
	out := CreateProviderVersionPlatformResponse{}
	if err = handleResponse(req, resp, &out, http.StatusCreated); err != nil {
		return nil, err
	}
	return &out, nil
}
