package get

import "github.com/go-playground/validator/v10"

type PVZDataRequest struct {
	StartDate string `query:"startDate" validate:"omitempty,datetime=2006-01-02T15:04:05Z07:00"`
	EndDate   string `query:"endDate" validate:"omitempty,datetime=2006-01-02T15:04:05Z07:00"`
	Page      int    `query:"page" validate:"omitempty,min=1"`
	Limit     int    `query:"limit" validate:"omitempty,min=1,max=30"`
}

func validateGetPVZDataRequest(req PVZDataRequest) error {
	v := validator.New()
	return v.Struct(req)
}
