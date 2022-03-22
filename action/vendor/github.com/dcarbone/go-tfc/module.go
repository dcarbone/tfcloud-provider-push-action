package tfc

import (
	"context"
	"net/http"
)

// ModuleClient interfaces with the Module Registry in Terraform Cloud
type ModuleClient struct {
	m *clientMiddleware
}

// ModuleClient interfaces with the Module Registry in Terraform Cloud
func (c *Client) ModuleClient() *ModuleClient {
	return &ModuleClient{m: c.m}
}

// CreateModuleVersion
//
// Executes: POST /api/v2/organizations/:organization_name/registry-modules/:registry-name/:namespace/:module_name/:provider_name/versions
// Docs:	 https://www.terraform.io/cloud-docs/api-docs/private-registry/modules#create-a-module-version
func (mc *ModuleClient) CreateModuleVersion(
	ctx context.Context,
	bearerToken,
	organizationName,
	registryName,
	namespace,
	moduleName,
	providerName string,
	data CreateModuleVersionRequest) (*CreateModuleVersionResponse, error) {
	route := buildRoute(
		pathAPI,
		pathV2,
		pathOrganizations,
		organizationName,
		pathRegistryModules,
		registryName,
		namespace,
		moduleName,
		providerName,
		pathVersions,
	)
	req, err := mc.m.buildRequest(ctx, http.MethodPost, route, nil, data)
	if err != nil {
		return nil, err
	}
	setBearerToken(req, bearerToken)
	req.Header.Set(headerContentType, applicationVNDAPIJSON)
	req.Header.Set(headerAccept, applicationJSON)
	resp, err := mc.m.do(req)
	if err != nil {
		return nil, err
	}
	out := CreateModuleVersionResponse{}
	if err = handleResponse(req, resp, &out, http.StatusCreated); err != nil {
		return nil, err
	}
	return &out, nil
}
