package handler

import (
	"net/http"

	"github.com/meraf00/swytch/core/lib/respond"
	"github.com/meraf00/swytch/core/lib/validation"
	"github.com/meraf00/swytch/internal/app"
)

// Generate pre-signed upload url
func HandleFileUpload(fs app.FileService) http.HandlerFunc {
	type uploadRequest struct {
		Filename string `json:"filename" validate:"required"`
	}

	type uploadResponse struct {
		UploadURL string `json:"upload_url"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		validator := validation.NewValidator(validation.ValidationSchemas{
			Body: &uploadRequest{},
		})

		body, err := validator.GetParams(r)
		if err != nil {
			respond.Error(w, err)
			return
		}

		req := body.(*uploadRequest)

		url, err := fs.GenerateUploadUrl(ctx, req.Filename)
		if err != nil {
			respond.Error(w, err)
		}

		respond.SuccessWithData(w, http.StatusOK, &uploadResponse{
			UploadURL: url.String(),
		})
	}
}
