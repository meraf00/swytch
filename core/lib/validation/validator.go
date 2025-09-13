package validation

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/go-playground/validator/v10"
	"github.com/gorilla/mux"

	appError "github.com/meraf00/swytch/core/lib/apperror"
)

const (
	ParamValidationError = "params_validation"
	BodyValidationError  = "body_validation"
	QueryValidationError = "query_validation"
)

type ValidationSchemas struct {
	Body    any
	Params  any
	Query   any
	Headers any
	Cookies any
}

type Validator struct {
	validate *validator.Validate
	schemas  ValidationSchemas
}

func NewValidator(schemas ValidationSchemas) *Validator {
	return &Validator{
		validate: validator.New(),
		schemas:  schemas,
	}
}

func (v *Validator) GetBody(r *http.Request) (any, error) {
	if v.schemas.Body == nil {
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			if errors.Is(err, io.EOF) {
				return map[string]any{}, nil
			}
			return nil, err
		}
		return body, nil
	}

	if err := json.NewDecoder(r.Body).Decode(v.schemas.Body); err != nil {
		if errors.Is(err, io.EOF) {
			return nil, appError.BadRequest("request body is required", BodyValidationError, nil)
		}
		return nil, err
	}

	if err := v.validate.Struct(v.schemas.Body); err != nil {
		if verr, ok := err.(validator.ValidationErrors); ok {
			return nil, appError.ValidationError(BodyValidationError, verr)
		}
		return nil, err
	}

	return v.schemas.Body, nil
}

func (v *Validator) GetParams(r *http.Request) (any, error) {
	vars := mux.Vars(r)
	if v.schemas.Params == nil {
		return vars, nil
	}

	data, err := json.Marshal(vars)
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(data, v.schemas.Params); err != nil {
		return nil, err
	}

	if err := v.validate.Struct(v.schemas.Params); err != nil {
		if verr, ok := err.(validator.ValidationErrors); ok {
			return nil, appError.ValidationError(ParamValidationError, verr)
		}
		return nil, err
	}

	return v.schemas.Params, nil
}

func (v *Validator) GetQuery(r *http.Request) (any, error) {
	query := make(map[string]any)
	q := r.URL.Query()

	for key, values := range q {
		if len(values) == 1 {
			query[key] = values[0]
		} else {
			query[key] = values
		}
	}

	if v.schemas.Query == nil {
		return query, nil
	}

	data, err := json.Marshal(query)
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(data, v.schemas.Query); err != nil {
		return nil, err
	}

	if err := v.validate.Struct(v.schemas.Query); err != nil {
		if verr, ok := err.(validator.ValidationErrors); ok {
			return nil, appError.ValidationError(QueryValidationError, verr)
		}
		return nil, err
	}

	return v.schemas.Query, nil
}

func (v *Validator) BindAndValidateQuery(r *http.Request, target any) error {
	q := r.URL.Query()
	query := make(map[string]any)

	for key, values := range q {
		if len(values) == 1 {
			query[key] = values[0]
		} else {
			query[key] = values
		}
	}

	data, err := json.Marshal(query)
	if err != nil {
		return err
	}

	if err := json.Unmarshal(data, target); err != nil {
		return err
	}

	if err := v.validate.Struct(target); err != nil {
		if verr, ok := err.(validator.ValidationErrors); ok {
			return appError.ValidationError(QueryValidationError, verr)
		}
		return err
	}

	return nil
}

func (v *Validator) BindAndValidateBody(r *http.Request, target any) error {
	if err := json.NewDecoder(r.Body).Decode(target); err != nil {
		return err
	}

	if err := v.validate.Struct(target); err != nil {
		if verr, ok := err.(validator.ValidationErrors); ok {
			return appError.ValidationError(BodyValidationError, verr)
		}
		return err
	}

	return nil
}

func (v *Validator) BindAndValidateParams(r *http.Request, target any) error {
	vars := mux.Vars(r)
	data, err := json.Marshal(vars)
	if err != nil {
		return err
	}

	if err := json.Unmarshal(data, target); err != nil {
		return err
	}

	if err := v.validate.Struct(target); err != nil {
		if verr, ok := err.(validator.ValidationErrors); ok {
			return appError.ValidationError(ParamValidationError, verr)
		}
		return err
	}

	return nil
}
