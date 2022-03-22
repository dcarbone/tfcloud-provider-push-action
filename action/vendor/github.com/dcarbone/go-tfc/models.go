package tfc

const (
	requestTypeRegistryProviderVersions         = "registry-provider-versions"
	requestTypeRegistryProviderVersionPlatforms = "registry-provider-version-platforms"
	requestTypeRegistryModuleVersions           = "registry-module-versions"
)

type (
	CreateProviderVersionRequestDataAttributes struct {
		Version   string   `json:"version"`
		KeyID     string   `json:"key-id"`
		Protocols []string `json:"protocols"`
	}
	CreateProviderVersionRequestData struct {
		Type       string                                     `json:"type"`
		Attributes CreateProviderVersionRequestDataAttributes `json:"attributes"`
	}

	CreateProviderVersionRequest struct {
		Data CreateProviderVersionRequestData `json:"data"`
	}
)

func NewCreateProviderVersionRequest(vers, keyID string, protocols []string) CreateProviderVersionRequest {
	m := CreateProviderVersionRequest{
		Data: CreateProviderVersionRequestData{
			Type: requestTypeRegistryProviderVersions,
			Attributes: CreateProviderVersionRequestDataAttributes{
				Protocols: protocols,
				Version:   vers,
				KeyID:     keyID,
			},
		},
	}
	return m
}

type (
	CreateProviderVersionResponseDataAttributesPermissions struct {
		CanDelete      bool `json:"can-delete"`
		CanUploadAsset bool `json:"can-upload-asset"`
	}

	CreateProviderVersionResponseDataAttributes struct {
		CreatedAt          string                                                 `json:"created-at"`
		KeyID              string                                                 `json:"key-id"`
		Permissions        CreateProviderVersionResponseDataAttributesPermissions `json:"permissions"`
		Protocols          []string                                               `json:"protocols"`
		ShasumsSigUploaded bool                                                   `json:"shasums-sig-uploaded"`
		ShasumsUploaded    bool                                                   `json:"shasums-uploaded"`
		UpdatedAt          string                                                 `json:"updated-at"`
		Version            string                                                 `json:"version"`
	}

	CreateProviderVersionResponseDataLinks struct {
		ShasumsSigUpload string `json:"shasums-sig-upload"`
		ShasumsUpload    string `json:"shasums-upload"`
	}

	CreateProviderVersionResponseDataRelationshipsRegistryProviderData struct {
		ID   string `json:"id"`
		Type string `json:"type"`
	}

	CreateProviderVersionResponseDataRelationshipsRegistryProvider struct {
		Data CreateProviderVersionResponseDataRelationshipsRegistryProviderData `json:"data"`
	}

	CreateProviderVersionResponseDataRelationshipsRegistryProviderPlatformsLinks struct {
		Related string `json:"related"`
	}

	CreateProviderVersionResponseDataRelationshipsRegistryProviderPlatforms struct {
		Data  []interface{}                                                                `json:"data"`
		Links CreateProviderVersionResponseDataRelationshipsRegistryProviderPlatformsLinks `json:"links"`
	}

	CreateProviderVersionResponseDataRelationships struct {
		RegistryProvider          CreateProviderVersionResponseDataRelationshipsRegistryProvider          `json:"registry-provider"`
		RegistryProviderPlatforms CreateProviderVersionResponseDataRelationshipsRegistryProviderPlatforms `json:"registry-provider-platforms"`
	}

	CreateProviderVersionResponseData struct {
		Attributes    CreateProviderVersionResponseDataAttributes    `json:"attributes"`
		ID            string                                         `json:"id"`
		Links         CreateProviderVersionResponseDataLinks         `json:"links"`
		Relationships CreateProviderVersionResponseDataRelationships `json:"relationships"`
		Type          string                                         `json:"type"`
	}

	CreateProviderVersionResponse struct {
		Data CreateProviderVersionResponseData `json:"data"`
	}
)

type (
	CreateProviderVersionPlatformRequestDataAttributes struct {
		Arch     string `json:"arch"`
		Filename string `json:"filename"`
		OS       string `json:"os"`
		Shasum   string `json:"shasum"`
	}

	CreateProviderVersionPlatformRequestData struct {
		Attributes CreateProviderVersionPlatformRequestDataAttributes `json:"attributes"`
		Type       string                                             `json:"type"`
	}

	CreateProviderVersionPlatformRequest struct {
		Data CreateProviderVersionPlatformRequestData `json:"data"`
	}
)

func NewCreateProviderVersionPlatformRequest(os, arch, shasum, filename string) CreateProviderVersionPlatformRequest {
	m := CreateProviderVersionPlatformRequest{
		Data: CreateProviderVersionPlatformRequestData{
			Type: requestTypeRegistryProviderVersionPlatforms,
			Attributes: CreateProviderVersionPlatformRequestDataAttributes{
				Arch:     arch,
				Filename: filename,
				OS:       os,
				Shasum:   shasum,
			},
		},
	}

	return m
}

type (
	CreateProviderVersionPlatformResponseDataAttributesPermissions struct {
		CanDelete      bool `json:"can-delete"`
		CanUploadAsset bool `json:"can-upload-asset"`
	}

	CreateProviderVersionPlatformResponseDataAttributes struct {
		Arch                   string                                                         `json:"arch"`
		Filename               string                                                         `json:"filename"`
		Os                     string                                                         `json:"os"`
		Permissions            CreateProviderVersionPlatformResponseDataAttributesPermissions `json:"permissions"`
		ProviderBinaryUploaded bool                                                           `json:"provider-binary-uploaded"`
		Shasum                 string                                                         `json:"shasum"`
	}

	CreateProviderVersionPlatformResponseDataLinks struct {
		ProviderBinaryUpload string `json:"provider-binary-upload"`
	}

	CreateProviderVersionPlatformResponseDataRelationshipsRegistryProviderVersionData struct {
		ID   string `json:"id"`
		Type string `json:"type"`
	}

	CreateProviderVersionPlatformResponseDataRelationshipsRegistryProviderVersion struct {
		Data CreateProviderVersionPlatformResponseDataRelationshipsRegistryProviderVersionData `json:"data"`
	}

	CreateProviderVersionPlatformResponseDataRelationships struct {
		RegistryProviderVersion CreateProviderVersionPlatformResponseDataRelationshipsRegistryProviderVersion `json:"registry-provider-version"`
	}

	CreateProviderVersionPlatformResponseData struct {
		Attributes    CreateProviderVersionPlatformResponseDataAttributes    `json:"attributes"`
		ID            string                                                 `json:"id"`
		Links         CreateProviderVersionPlatformResponseDataLinks         `json:"links"`
		Relationships CreateProviderVersionPlatformResponseDataRelationships `json:"relationships"`
		Type          string                                                 `json:"type"`
	}

	CreateProviderVersionPlatformResponse struct {
		Data CreateProviderVersionPlatformResponseData `json:"data"`
	}
)

type (
	CreateModuleVersionRequestDataAttributes struct {
		Version string `json:"version"`
	}

	CreateModuleVersionRequestData struct {
		Attributes CreateModuleVersionRequestDataAttributes `json:"attributes"`
		Type       string                                   `json:"type"`
	}

	CreateModuleVersionRequest struct {
		Data CreateModuleVersionRequestData `json:"data"`
	}
)

func NewCreateModuleVersionRequest(version string) CreateModuleVersionRequest {
	m := CreateModuleVersionRequest{
		Data: CreateModuleVersionRequestData{
			Type: requestTypeRegistryModuleVersions,
			Attributes: CreateModuleVersionRequestDataAttributes{
				Version: version,
			},
		},
	}

	return m
}

type (
	CreateModuleVersionResponseDataAttributes struct {
		CreatedAt string `json:"created-at"`
		Source    string `json:"source"`
		Status    string `json:"status"`
		UpdatedAt string `json:"updated-at"`
		Version   string `json:"version"`
	}

	CreateModuleVersionResponseDataLinks struct {
		Upload string `json:"upload"`
	}

	CreateModuleVersionResponseDataRelationshipsRegistryModuleData struct {
		ID   string `json:"id"`
		Type string `json:"type"`
	}

	CreateModuleVersionResponseDataRelationshipsRegistryModule struct {
		Data CreateModuleVersionResponseDataRelationshipsRegistryModuleData `json:"data"`
	}

	CreateModuleVersionResponseDataRelationships struct {
		RegistryModule CreateModuleVersionResponseDataRelationshipsRegistryModule `json:"registry-module"`
	}

	CreateModuleVersionResponseData struct {
		Attributes    CreateModuleVersionResponseDataAttributes    `json:"attributes"`
		ID            string                                       `json:"id"`
		Links         CreateModuleVersionResponseDataLinks         `json:"links"`
		Relationships CreateModuleVersionResponseDataRelationships `json:"relationships"`
		Type          string                                       `json:"type"`
	}

	CreateModuleVersionResponse struct {
		Data CreateModuleVersionResponseData `json:"data"`
	}
)
