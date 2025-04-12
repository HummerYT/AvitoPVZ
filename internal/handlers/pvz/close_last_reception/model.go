package close_last_reception

import "github.com/go-playground/validator/v10"

type CloseReceptionRequest struct {
	PvzID string `param:"pvzId" validate:"required,uuid"`
}

func validateCloseReceptionRequest(req CloseReceptionRequest) error {
	v := validator.New()
	return v.Struct(req)
}
