package main

import (
	"encoding/json"
	"fmt"
	"strings"
)

const (
	tfRequestTypeRegistryProviderVersions         = "registry-provider-versions"
	tfRequestTypeRegistryProviderVersionPlatforms = "registry-provider-version-platforms"
)

type TFCloudAPIErrorError struct {
	Status string                 `json:"status"`
	Title  string                 `json:"title"`
	Detail string                 `json:"detail"`
	Source map[string]interface{} `json:"source"`
}

type TFCloudAPIError struct {
	Errors  []TFCloudAPIErrorError `json:"errors,omitempty"`
	Success *bool                  `json:"success,omitempty"`
}

func (e *TFCloudAPIError) UnmarshalJSON(b []byte) error {
	type tmp struct {
		Success *bool           `json:"success"`
		Errors  json.RawMessage `json:"errors"`
	}

	t := new(tmp)
	if err := json.Unmarshal(b, t); err != nil {
		return err
	}

	// "errors"" is usually an array of objects, but because javascript and other horrible things exist, it can
	// sometimes be an array of strings.

	e.Errors = make([]TFCloudAPIErrorError, 0)

	// test to see if its a friggin array of friggin strings.
	if strings.HasPrefix(string(t.Errors), "{\"errors\":[\"") {
		strs := make([]string, 0)
		if err := json.Unmarshal(t.Errors, &strs); err != nil {
			return err
		}
		for _, str := range strs {
			e.Errors = append(e.Errors, TFCloudAPIErrorError{Title: str})
		}
		return nil
	}

	e.Errors = make([]TFCloudAPIErrorError, 0)
	return json.Unmarshal(t.Errors, &e.Errors)
}

type StatusError struct {
	ExpectedCode int             `json:"expected-code"`
	ActualCode   int             `json:"actual-code"`
	Body         string          `json:"body"`
	TFCloudError TFCloudAPIError `json:"tfcloud-error"`
}

func (e StatusError) Error() string {
	return fmt.Sprintf("expected response code %v, saw: code=%d; body=\"%s\"", e.ExpectedCode, e.ActualCode, e.Body)
}

type (
	TFCreateProviderVersionRequestDataAttributes struct {
		Version   string   `json:"version"`
		KeyID     string   `json:"key-id"`
		Protocols []string `json:"protocols"`
	}
	TFCreateProviderVersionRequestData struct {
		Type       string                                       `json:"type"`
		Attributes TFCreateProviderVersionRequestDataAttributes `json:"attributes"`
	}

	TFCreateProviderVersionRequest struct {
		Data TFCreateProviderVersionRequestData `json:"data"`
	}
)

func NewTFCreateProviderVersionRequest(vers, keyID string, protocols []string) TFCreateProviderVersionRequest {
	m := TFCreateProviderVersionRequest{
		Data: TFCreateProviderVersionRequestData{
			Type: tfRequestTypeRegistryProviderVersions,
			Attributes: TFCreateProviderVersionRequestDataAttributes{
				Protocols: protocols,
				Version:   vers,
				KeyID:     keyID,
			},
		},
	}
	return m
}

type (
	TFCreateProviderVersionResponseDataAttributesPermissions struct {
		CanDelete      bool `json:"can-delete"`
		CanUploadAsset bool `json:"can-upload-asset"`
	}

	TFCreateProviderVersionResponseDataAttributes struct {
		CreatedAt          string                                                   `json:"created-at"`
		KeyID              string                                                   `json:"key-id"`
		Permissions        TFCreateProviderVersionResponseDataAttributesPermissions `json:"permissions"`
		Protocols          []string                                                 `json:"protocols"`
		ShasumsSigUploaded bool                                                     `json:"shasums-sig-uploaded"`
		ShasumsUploaded    bool                                                     `json:"shasums-uploaded"`
		UpdatedAt          string                                                   `json:"updated-at"`
		Version            string                                                   `json:"version"`
	}

	TFCreateProviderVersionResponseDataLinks struct {
		ShasumsSigUpload string `json:"shasums-sig-upload"`
		ShasumsUpload    string `json:"shasums-upload"`
	}

	TFCreateProviderVersionResponseDataRelationshipsRegistryProviderData struct {
		ID   string `json:"id"`
		Type string `json:"type"`
	}

	TFCreateProviderVersionResponseDataRelationshipsRegistryProvider struct {
		Data TFCreateProviderVersionResponseDataRelationshipsRegistryProviderData `json:"data"`
	}

	TFCreateProviderVersionResponseDataRelationshipsRegistryProviderPlatformsLinks struct {
		Related string `json:"related"`
	}

	TFCreateProviderVersionResponseDataRelationshipsRegistryProviderPlatforms struct {
		Data  []interface{}                                                                  `json:"data"`
		Links TFCreateProviderVersionResponseDataRelationshipsRegistryProviderPlatformsLinks `json:"links"`
	}

	TFCreateProviderVersionResponseDataRelationships struct {
		RegistryProvider          TFCreateProviderVersionResponseDataRelationshipsRegistryProvider          `json:"registry-provider"`
		RegistryProviderPlatforms TFCreateProviderVersionResponseDataRelationshipsRegistryProviderPlatforms `json:"registry-provider-platforms"`
	}

	TFCreateProviderVersionResponseData struct {
		Attributes    TFCreateProviderVersionResponseDataAttributes    `json:"attributes"`
		ID            string                                           `json:"id"`
		Links         TFCreateProviderVersionResponseDataLinks         `json:"links"`
		Relationships TFCreateProviderVersionResponseDataRelationships `json:"relationships"`
		Type          string                                           `json:"type"`
	}

	TFCreateProviderVersionResponse struct {
		Data TFCreateProviderVersionResponseData `json:"data"`
	}
)

type (
	TFCreateProviderVersionPlatformRequestDataAttributes struct {
		Arch     string `json:"arch"`
		Filename string `json:"filename"`
		OS       string `json:"os"`
		Shasum   string `json:"shasum"`
	}

	TFCreateProviderVersionPlatformRequestData struct {
		Attributes TFCreateProviderVersionPlatformRequestDataAttributes `json:"attributes"`
		Type       string                                               `json:"type"`
	}

	TFCreateProviderVersionPlatformRequest struct {
		Data TFCreateProviderVersionPlatformRequestData `json:"data"`
	}
)

func NewTFCreateProviderVersionPlatformRequest(os, arch, shasum, filename string) TFCreateProviderVersionPlatformRequest {
	m := TFCreateProviderVersionPlatformRequest{
		Data: TFCreateProviderVersionPlatformRequestData{
			Type: tfRequestTypeRegistryProviderVersionPlatforms,
			Attributes: TFCreateProviderVersionPlatformRequestDataAttributes{
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
	TFCreateProviderVersionPlatformResponseDataAttributesPermissions struct {
		CanDelete      bool `json:"can-delete"`
		CanUploadAsset bool `json:"can-upload-asset"`
	}

	TFCreateProviderVersionPlatformResponseDataAttributes struct {
		Arch                   string                                                           `json:"arch"`
		Filename               string                                                           `json:"filename"`
		Os                     string                                                           `json:"os"`
		Permissions            TFCreateProviderVersionPlatformResponseDataAttributesPermissions `json:"permissions"`
		ProviderBinaryUploaded bool                                                             `json:"provider-binary-uploaded"`
		Shasum                 string                                                           `json:"shasum"`
	}

	TFCreateProviderVersionPlatformResponseDataLinks struct {
		ProviderBinaryUpload string `json:"provider-binary-upload"`
	}

	TFCreateProviderVersionPlatformResponseDataRelationshipsRegistryProviderVersionData struct {
		ID   string `json:"id"`
		Type string `json:"type"`
	}

	TFCreateProviderVersionPlatformResponseDataRelationshipsRegistryProviderVersion struct {
		Data TFCreateProviderVersionPlatformResponseDataRelationshipsRegistryProviderVersionData `json:"data"`
	}

	TFCreateProviderVersionPlatformResponseDataRelationships struct {
		RegistryProviderVersion TFCreateProviderVersionPlatformResponseDataRelationshipsRegistryProviderVersion `json:"registry-provider-version"`
	}

	TFCreateProviderVersionPlatformResponseData struct {
		Attributes    TFCreateProviderVersionPlatformResponseDataAttributes    `json:"attributes"`
		ID            string                                                   `json:"id"`
		Links         TFCreateProviderVersionPlatformResponseDataLinks         `json:"links"`
		Relationships TFCreateProviderVersionPlatformResponseDataRelationships `json:"relationships"`
		Type          string                                                   `json:"type"`
	}

	TFCreateProviderVersionPlatformResponse struct {
		Data TFCreateProviderVersionPlatformResponseData `json:"data"`
	}
)
