package tfc

import (
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/hashicorp/go-cleanhttp"
)

var UploadClient *http.Client

func init() {
	UploadClient = cleanhttp.DefaultClient()
}

type FileUploadRequest struct {
	File        io.Reader
	Destination string
	ContentType string
	Filename    string
}

// UploadArtifact wraps the basics around uploading an artifact to Terraform Cloud
func UploadArtifact(ctx context.Context, data FileUploadRequest) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodPut, data.Destination, data.File)
	if err != nil {
		return fmt.Errorf("error constructing request: %w", err)
	}
	req.Header.Set(headerContentType, data.ContentType)
	req.Header.Set(headerContentDisposition, fmt.Sprintf(attachmentFilenameFmt, data.Filename))
	resp, err := UploadClient.Do(req)
	if err != nil {
		return err
	}
	return handleResponse(req, resp, nil, http.StatusOK)
}
