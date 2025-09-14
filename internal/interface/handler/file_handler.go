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
		UploadUrl string `json:"upload_url"`
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
			UploadUrl: url.RawPath,
		})
	}
}

// Generate pre-signed download url for completed task
func HandleTaskDownload(cs *app.ConversionService) http.HandlerFunc {
	type downloadRequest struct {
		TaskID string `json:"task_id" validate:"required"`
	}

	type response struct {
		DownloadUrl string `json:"download_url"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		validator := validation.NewValidator(validation.ValidationSchemas{
			Params: &downloadRequest{},
		})

		body, err := validator.GetParams(r)
		if err != nil {
			respond.Error(w, err)
			return
		}

		req := body.(*downloadRequest)

		url, err := cs.GenerateTaskDownloadUrl(ctx, req.TaskID)
		if err != nil {
			respond.Error(w, err)
			return
		}

		respond.JSON(w, http.StatusOK, &response{
			DownloadUrl: url.RawPath,
		})
	}
}
