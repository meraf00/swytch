package handler

import (
	"net/http"
	"time"

	"github.com/meraf00/swytch/core/lib/respond"
	"github.com/meraf00/swytch/core/lib/validation"
	"github.com/meraf00/swytch/internal/app"
)

// Get job by ID
func HandleGetJob(cs *app.ConversionService) http.HandlerFunc {
	type getJobRequest struct {
		ID string `json:"job_id" validate:"required"`
	}

	type response struct {
		JobID     string    `json:"job_id"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		validator := validation.NewValidator(validation.ValidationSchemas{
			Params: &getJobRequest{},
		})

		body, err := validator.GetParams(r)
		if err != nil {
			respond.Error(w, err)
			return
		}

		req := body.(*getJobRequest)

		job, err := cs.GetJob(ctx, req.ID)

		if err != nil {
			respond.Error(w, err)
			return
		}

		respond.JSON(w, http.StatusOK, &response{
			JobID:     job.ID,
			CreatedAt: job.CreatedAt,
			UpdatedAt: job.UpdatedAt,
		})
	}
}

// Get tasks for a job
func HandleGetJobTasks(cs *app.ConversionService) http.HandlerFunc {
	type getJobRequest struct {
		ID string `json:"job_id" validate:"required"`
	}

	type responseTask struct {
		ID                string    `json:"id"`
		Status            string    `json:"status"`
		CreatedAt         time.Time `json:"created_at"`
		UpdatedAt         time.Time `json:"updated_at"`
		ObjectName        string    `json:"object_name"`
		OriginalName      string    `json:"original_name"`
		OriginalFormat    string    `json:"original_format"`
		TargetFormat      string    `json:"target_format"`
		ConvertedFileName string    `json:"converted_file_name,omitempty"`
	}

	type response struct {
		Tasks []responseTask `json:"tasks"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		validator := validation.NewValidator(validation.ValidationSchemas{
			Params: &getJobRequest{},
		})

		body, err := validator.GetParams(r)
		if err != nil {
			respond.Error(w, err)
			return
		}

		req := body.(*getJobRequest)

		tasks, err := cs.GetJobTasks(ctx, req.ID)

		if err != nil {
			respond.Error(w, err)
			return
		}

		res := make([]responseTask, len(tasks))

		for i, task := range tasks {
			res[i] = responseTask{
				ID:                task.ID,
				Status:            string(task.Status),
				CreatedAt:         task.CreatedAt,
				UpdatedAt:         task.UpdatedAt,
				ObjectName:        task.File.ObjectName,
				OriginalName:      task.File.OriginalName,
				OriginalFormat:    task.File.OriginalFormat,
				TargetFormat:      task.TargetFormat,
				ConvertedFileName: task.ConvertedFileName,
			}
		}

		respond.JSON(w, http.StatusOK, &response{
			Tasks: res,
		})
	}
}

// Create a new conversion job with related tasks and files
func HandleCreateJob(cs *app.ConversionService) http.HandlerFunc {
	type createJobRequest struct {
		Files []struct {
			ObjectName     string   `json:"object_name"`
			OriginalName   string   `json:"original_name"`
			OriginalFormat string   `json:"original_format"`
			TargetFormats  []string `json:"target_formats"`
		} `json:"files"`
	}

	type response struct {
		JobID       string `json:"job_id"`
		ProgressURL string `json:"progress_url"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		validator := validation.NewValidator(validation.ValidationSchemas{
			Body: &createJobRequest{},
		})

		body, err := validator.GetBody(r)
		if err != nil {
			respond.Error(w, err)
			return
		}

		req := body.(*createJobRequest)

		jobID, err := cs.CreateJob(ctx, &app.CreateJobParams{
			Files: []struct {
				ObjectName     string
				OriginalName   string
				OriginalFormat string
				TargetFormats  []string
			}(req.Files),
		})

		if err != nil {
			respond.Error(w, err)
			return
		}

		respond.JSON(w, http.StatusOK, &response{
			JobID:       jobID,
			ProgressURL: "/jobs/" + jobID + "/status",
		})
	}
}

// Generate pre-signed download url for completed task
func HandleGetCompletedTaskDownloadURL(cs *app.ConversionService) http.HandlerFunc {
	type downloadRequest struct {
		TaskID string `json:"task_id" validate:"required"`
	}

	type response struct {
		DownloadURL string `json:"download_url"`
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
			DownloadURL: url.String(),
		})
	}
}
